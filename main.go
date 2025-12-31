package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"VPNClient/vna"

	"github.com/joho/godotenv"
)

func main() {
	////root context 
	rootCtx, rootCancel := context.WithCancel(context.Background())
	
	_ = godotenv.Load() 

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

	////=============CREATE VNA OBJECT=============
	vna, err := vna.New(rootCtx,"vpn0","192.168.0.50:5000",":5000")	
	
	if err != nil {
		log.Fatalf("nepovedlo se vytvořit VNA: %v", err)
	}
	
	////stop lifecycle
    defer vna.Stop()
	

	////=============BOOT VNA=============
	if err := vna.Boot(); err != nil{
		fmt.Println(err)
		rootCancel();
	}


	//// ===========START LIFECYCLE=============
	vna.Start()	
	
	log.Println("VNA running — press Ctrl+C for end")
    <-rootCtx.Done()
    log.Println("main exiting")

}
