package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"VPNClient/vna"
)

func main() {
	////root context 

	rootCtx, rootCancel := context.WithCancel(context.Background())

	defer rootCancel()

	////signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	
	////goruntine checking signal for end
	go func ()  {
		<- sigs
		log.Println("signal recieved, canceling root context")
		rootCancel()
	}()

	ifName := "MyVirtualAdapter"
	ip := "10.0.0.1"
	mask := "255.255.255.0"

	///buffer size 2 Mb
	var bufferSize uint32 = 0x200000 

	////Create Virtual Network Adapter
	vna, err := vna.New(rootCtx,ifName,ip,mask, bufferSize)

	if err != nil {
		log.Fatalf("nepovedlo se vytvořit VNA: %v", err)
	}

    defer vna.Stop()
	
	vna.Start()

	if err := vna.SetupAdapter(); err != nil{
		fmt.Println(err)
		rootCancel();
	}

	log.Println("VNA běží — stiskni Ctrl+C pro ukončení")
    <-rootCtx.Done()
    log.Println("main exiting")

}
