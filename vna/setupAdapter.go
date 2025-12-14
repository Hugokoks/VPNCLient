package vna

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"VPNClient/services"
)

// RemoveRoute removes the default route for this adapter.
func (v *VNA) RemoveRoute() error {

    ctxDel, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	
    delCmd := []string{
        "netsh", "interface", "ipv4", "delete", "route",
        "0.0.0.0/0", strconv.Itoa(v.AdapterIndex),
    }
	
	/*
	////delete testing route
	delCmd := []string{
        "netsh", "interface", "ipv4", "delete", "route",
        "10.0.0.0/24", strconv.Itoa(v.AdapterIndex),
    }
		*/
    return services.RunCommand(ctxDel, delCmd)
}

// getAdapterIndexFromNetsh parses the index from netsh output
func (v * VNA) getAdapterIndexFromNetsh(output, ifName string)  error {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ifName) {
			fields := strings.Fields(line)
			if len(fields) < 1 {
				return fmt.Errorf("could not parse interface row: %q", line)
			}

			idx, err := strconv.Atoi(fields[0])
			if err != nil {
				return  fmt.Errorf("could not convert index: %w", err)
			}
			
			v.AdapterIndex = idx

			return  nil
		}
	}
	return  fmt.Errorf("interface %q not found", ifName)
}

func (v *VNA) SetupAdapter() error {

	//10s timeout for running netsh commands
	ctx, cancel := context.WithTimeout(v.ctx, 10*time.Second)
	defer cancel()

	nameArg := fmt.Sprintf("name=%q", v.IfName)

	// ============ SETUP IP =============
	cmdIP := []string{
		"netsh", "interface", "ip", "set", "address",
		nameArg, "static", v.IP, v.Mask,
	}

	if err := services.RunCommand(ctx, cmdIP); err != nil {
		return fmt.Errorf("setup IP failed: %w", err)
	}

	// ============ GET ADAPTER INDEX ============
	cmdShow := []string{"netsh", "interface", "ipv4", "show", "interfaces"}
	out, err := services.RunCommandWithOutput(ctx, cmdShow)

	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}

	if err := v.getAdapterIndexFromNetsh(out, v.IfName); err != nil{
		
		return fmt.Errorf("faild to get index of adaptet %w",&err)
	}

	// ============ SET DEFAULT ROUTE ============
	cmdRoute := []string{
		"netsh", "interface", "ipv4", "add", "route",
		"0.0.0.0/0", strconv.Itoa(v.AdapterIndex),"metric=1",
	}
	
	////Tunel test for IP 
	/* cmdRoute := []string{"netsh","interface","ipv4","add","route","10.0.0.0/24",strconv.Itoa(v.AdapterIndex),v.IP}*/
	if err := services.RunCommand(ctx, cmdRoute); err != nil {
		return fmt.Errorf("setup default route failed: %w", err)
	}


	///=============MTU SETTINGS===========================
	cmdMTU := []string{
    	"netsh", "interface", "ipv4", "set", "subinterface",
    	fmt.Sprintf("%q", v.IfName), "mtu=1400", "store=persistent",
	}

	if err := services.RunCommand(ctx, cmdMTU); err != nil{

		return fmt.Errorf("setup MTU failed %w",err)
	}

	/*
	///============FIREWALL SETTINGS======================
	if err := setVPNPrivate(); err != nil {
	    log.Printf("Varování: Nepodařilo se nastavit vpn0 jako Private: %v", err)
	}
	if err := allowICMPAndVPNTraffic(); err != nil {
	    log.Printf("Varování: Nepodařilo se nastavit firewall pravidla: %v", err)
	}
	/*
	cmd := exec.Command("netsh", "advfirewall", "set", "allprofiles", "state", "off")
	if out, err := cmd.CombinedOutput(); err != nil {
    	log.Printf("Nepodařilo se vypnout firewall: %v (výstup: %s)", err, string(out))
	} else {
	    log.Println("=== FIREWALL DOČASNĚ VYPNUT PRO TEST STABILITY ===")
	    log.Println("Po testu ho zapni ručně: netsh advfirewall set allprofiles state on")
	}
	*/

	// ============ CREATE CryptoKeys Struct and load it in vna ============
	if err := v.LoadCryptoKeys();err != nil{
		return fmt.Errorf("key laod: %w", err) 
	}
	
	// ============ SET CONNECTION ============
	if err := v.InitConnection(); err != nil{

		return  fmt.Errorf("connection setup failed %v",err)

	}

	fmt.Println("Connection successully set")

	// ============ HANDSHAKE WITH SERVER ============
	if err := v.Handshake(); err!=nil{

		return fmt.Errorf("handshake failed: %w",err)
	}

	fmt.Println("VNA successfully settled!")
	
	return nil
}


// Nastaví síťový profil vpn0 na Private
func setVPNPrivate() error {
    // Počkáme chvíli, aby se adaptér objevil v systému
    time.Sleep(2 * time.Second)

    // Získáme název rozhraní přes PowerShell
    cmd := exec.Command("powershell", "-Command",
        `Get-NetConnectionProfile | Where-Object {$_.InterfaceAlias -like "*vpn0*"} | Select-Object -ExpandProperty InterfaceIndex`)

    out, err := cmd.Output()
    if err != nil {
        return err
    }

    interfaceIndex := strings.TrimSpace(string(out))
    if interfaceIndex == "" {
        return fmt.Errorf("vpn0 rozhraní nenalezeno")
    }

    // Nastavíme na Private
    cmd2 := exec.Command("powershell", "-Command",
        fmt.Sprintf(`Set-NetConnectionProfile -InterfaceIndex %s -NetworkCategory Private`, interfaceIndex))

    if err := cmd2.Run(); err != nil {
        return err
    }

    log.Println("vpn0 nastaveno jako Private network")
    return nil
}

// Povolí základní pravidla firewallu pro VPN
func allowICMPAndVPNTraffic() error {
    rules := []string{
        // Povolí příchozí ICMP echo reply (pro ping)
        `netsh advfirewall firewall add rule name="VPN - Allow ICMPv4-In" protocol=icmpv4:8,any dir=in action=allow`,

        // Povolí veškerý příchozí provoz na vpn0 rozhraní (bezpečné, protože je to jen z tunelu)
        `netsh advfirewall firewall add rule name="VPN - Allow All In on vpn0" dir=in action=allow interface="vpn0"`,

        // Povolí veškerý odchozí provoz z vpn0
        `netsh advfirewall firewall add rule name="VPN - Allow All Out on vpn0" dir=out action=allow interface="vpn0"`,
    }

    for _, r := range rules {
        parts := strings.Fields(r)
        cmd := exec.Command(parts[0], parts[1:]...)
        if out, err := cmd.CombinedOutput(); err != nil {
            // Některá pravidla už mohou existovat → ignorujeme chybu
            if !strings.Contains(string(out), "already exists") {
                log.Printf("Firewall pravidlo selhalo: %s → %v", r, err)
            }
        }
    }

    log.Println("Firewall pravidla pro VPN nastavena")
    return nil
}
