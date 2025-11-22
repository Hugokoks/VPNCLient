package vna

import (
	"context"
	"fmt"
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
        "0.0.0.0/0", strconv.Itoa(v.AdapterIndex), v.IP,
    }

    return services.RunCommand(ctxDel, delCmd)
}

// getAdapterIndexFromNetsh parses the index from netsh output
func getAdapterIndexFromNetsh(output, ifName string) (int, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ifName) {
			fields := strings.Fields(line)
			if len(fields) < 1 {
				return 0, fmt.Errorf("could not parse interface row: %q", line)
			}

			idx, err := strconv.Atoi(fields[0])
			if err != nil {
				return 0, fmt.Errorf("could not convert index: %w", err)
			}
			return idx, nil
		}
	}
	return 0, fmt.Errorf("interface %q not found", ifName)
}

// SetupAdapter configures IP + default route using fields stored in the VNA struct.
func (v *VNA) SetupAdapter() error {

	// 1) 10s timeout for running netsh commands
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

	idx, err := getAdapterIndexFromNetsh(out, v.IfName)
	if err != nil {
		return err
	}
	v.AdapterIndex = idx


	// ============ SET DEFAULT ROUTE ============
	cmdRoute := []string{
		"netsh", "interface", "ipv4", "add", "route",
		"0.0.0.0/0", strconv.Itoa(v.AdapterIndex), v.IP, "metric=1",
	}

	if err := services.RunCommand(ctx, cmdRoute); err != nil {
		return fmt.Errorf("setup default route failed: %w", err)
	}



	return nil
}
