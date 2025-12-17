package vna

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"VPNClient/services"
)

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

	// ============ CREATE CryptoKeys Struct and load it in vna ============
	if err := v.LoadCryptoKeys();err != nil{
		return fmt.Errorf("key laod: %w", err) 
	}
	
	// ============ SET UDP CONNECTION ============
	if err := v.InitConnection(); err != nil{

		return  fmt.Errorf("connection setup failed %v",err)

	}

	fmt.Println("VNA successfully settled!")
	
	return nil
}


// RemoveRoute removes the default route for this adapter.
func (v *VNA) RemoveRoute() error {

    ctxDel, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	
    delCmd := []string{
        "netsh", "interface", "ipv4", "delete", "route",
        "0.0.0.0/0", strconv.Itoa(v.AdapterIndex),
    }
	
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

///create UDP connection with server
func (v *VNA) InitConnection() error {

	///Client Address
	laddr, err := net.ResolveUDPAddr("udp", v.LocalAddr)
	if err != nil {
		return err
	}

	///Server Address
	raddr, err := net.ResolveUDPAddr("udp", v.RemoteAddr)
	if err != nil {
		return err
	}

	////UDP Connection between Client - Server
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}

	v.Conn = conn
	return nil

}

////load crypto keys objet into vna object
func (v *VNA )LoadCryptoKeys()  error {
	
	serverPubB64 := os.Getenv("SERVER_PUBLIC_KEY")
	if serverPubB64 == "" {
		return  fmt.Errorf("SERVER_PUBLIC_KEY není nastaven v .env")
	}

	pubBytes, err := base64.StdEncoding.DecodeString(serverPubB64)
	if err != nil {
		return fmt.Errorf("SERVER_PUBLIC_KEY: špatný base64 formát: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize { 
		return fmt.Errorf("SERVER_PUBLIC_KEY: musí být %d bytů, má %d", ed25519.PublicKeySize, len(pubBytes))
	}

	cryptoKeys := CryptoKeys{
		ServerPub: ed25519.PublicKey(pubBytes), 
		SharedKey: nil,                         
	}

	v.Keys = cryptoKeys

	return nil
}