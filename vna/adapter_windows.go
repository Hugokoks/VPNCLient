package vna

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"VPNClient/services"
)

/*
	Create virutal network adapter in OS
	Set default route for virtual network adapter
	Set up MTU
	Delete default route when vna is deleted
*/

func (v *VNA) SetupAdapter() error {

	//10s timeout for running netsh commands
	ctx, cancel := context.WithTimeout(v.ctx, 10*time.Second)
	defer cancel()

	// ============ SETUP IP =============
	nameArg := fmt.Sprintf("name=%q", v.IfName)

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



