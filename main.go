package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"VPNClient/services"
	"VPNClient/vna"
)

func main() {
	// jméno interface
	ifName := "MyVirtualAdapter"

	// bufferSize pro StartSession - zkus 64kB (uprav podle potřeby)
	var bufferSize uint32 = 0x200000 // 2 MB — doporučeno Wintunem

	// 1) vytvoř VNA
	v, err := vna.New(ifName, bufferSize)
	if err != nil {
		log.Fatalf("nepovedlo se vytvořit VNA: %v", err)
	}
	// zajistíme zavření při ukončení mainu
	defer func() {
		// Close čeká na ukončení gorutin a zavře adapter
		v.Close()
	}()

	// 2) spust listener (handler běží v goroutině)
	v.RunListener(func(pkt []byte) {
		// Pozor: pkt může být sdílený buffer z wintun, pokud ho chceš zpracovat asynchronně,
		// musíš si ho zkopírovat. Tady pouze vypíšeme délku.
		// Pokud chceš zpracovávat mimo tuto gorutinu, udělej copy := append([]byte(nil), pkt...)
		fmt.Printf("přijato %d bytů\n", len(pkt))
	})

	// 3) nastav IP na adapteru (netsh vyžaduje admin práva)
	ip := "10.0.0.1"
	mask := "255.255.255.0"

	if err := services.SetupIP(ifName, ip, mask); err != nil {
		// pokud netsh selže, vypíšeme chybu, ale listener stále běží (můžeš se rozhodnout ukončit)
		log.Printf("SetupIP selhalo: %v", err)
	} else {
		log.Printf("Nastavena IP %s/%s na rozhraní %q", ip, mask, ifName)
	}


	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	
	log.Println("VNA běží — stiskni Ctrl+C pro ukončení")
	
	// čekej na signál
	<-sigs
	log.Println("shutdown...")
	
	v.Close()
}
