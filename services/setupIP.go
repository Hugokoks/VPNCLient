package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// SetupIP nastaví statickou IP na rozhraní ifName.
// ifName může obsahovat mezery.
func SetupIP(ifName, ip, mask string) error {
	// Timeout pro případ, že netsh zamrzne
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Netsh chce argument ve tvaru name=<name>
	// Pokud ifName může mít mezery, obalíme ho do uvozovek.
	nameArg := fmt.Sprintf("name=%q", ifName)

	cmd := exec.CommandContext(ctx, "netsh",
		"interface", "ip", "set", "address",
		nameArg, "static", ip, mask)


	////chytame errory a vystupy z commandu abychom je mohli nakonci programu vypsat 
	var out bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		// vrátíme i výstup netsh pro snadnější debug
		return fmt.Errorf("netsh failed: %v: %s", err, out.String())
	}
	return nil
}
