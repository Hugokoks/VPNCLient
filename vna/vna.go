package vna

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.zx2c4.com/wintun"
)

type VNA struct {
	Iface   *wintun.Adapter
	Session wintun.Session
	IfName    string
	IP 		string
	Mask    string
	AdapterIndex int
	ctx     context.Context
	cancel  context.CancelFunc
	closeOnce sync.Once



	wg sync.WaitGroup // producers (listeners)

	PacketChan chan []byte
}

func New(rootCtx context.Context, ifName string,ip string,mask string, bufferSize uint32) (*VNA, error) {

	///create virtual network interface
	iface, err := wintun.CreateAdapter(ifName, "Wintun", nil)

	if err != nil {
		return nil, fmt.Errorf("CreateAdapter: %w", err)
	}

	////start session
	sess, err := iface.StartSession(bufferSize)
	if err != nil {
		iface.Close()
		return nil, fmt.Errorf("StartSession: %w", err)
	}

    ctx, cancel := context.WithCancel(rootCtx)

	return &VNA{
		Iface:      iface,
		Session:    sess,
		IfName:     ifName,
		IP: 		ip,
		Mask: 		mask,
		ctx:        ctx,
		cancel:     cancel,
		PacketChan: make(chan []byte, 5000),
	}, nil
}

// RunListener spustí goroutine, která čte pakety a předává je do handleru.
// handler musí do sebe kopírovat data pokud je bude chtít zpracovat asynchronně.
func (v *VNA) RunListener() {
	v.wg.Add(1)
	go func() {
		defer v.wg.Done()
		for {
			// nejprve kontrola cancelu - rychlejší ukončení
			select {
			case <-v.ctx.Done():
				return
			default:
			}

			packet, err := v.Session.ReceivePacket()
			if err != nil {
				select {
				case <-v.ctx.Done():
					return // Ukončení při žádosti
				default:
					time.Sleep(50 * time.Millisecond) // Pauza, aby se nezanikl CPU
					continue                          // Opakovat cyklus
				}
			}

			copyPkt := append([]byte(nil), packet...)

			select {
			case v.PacketChan <- copyPkt:
			default:
			}

			// uvolnit buffer
			v.Session.ReleaseReceivePacket(packet)
		}
	}()
}
func (v *VNA) RunEncryptor() {
	v.wg.Add(1)
	go func() {
		defer v.wg.Done()

		var packetCount int
		var lastLog = time.Now()

		// TATO SMYČKA ČTE DATA A TÍM UVOLŇUJE KANÁL
		for rawPacket := range v.PacketChan {

			// Logika detekce ICMP zůstává stejná, ale nyní se provede
			if len(rawPacket) >= 20 && rawPacket[9] == 1 {
				log.Println("ICMP packet recieved.")
			}

			// Zde by normálně probíhalo šifrování a odeslání do sítě.
			// time.Sleep(1 * time.Microsecond) // Můžete přidat pro simulaci zátěže

			packetCount++
			if time.Since(lastLog) >= 1*time.Second {
				log.Printf("[Encryptor] Speed: %d packet:", packetCount)
				packetCount = 0
				lastLog = time.Now()
			}
		}
		log.Println("[Encryptor] ended.")
	}()
}


func (v *VNA) Start(){

	v.RunListener()
	v.RunEncryptor()

}
func (v *VNA) Stop(){

	v.Close()
}

func (v *VNA) Close() {
	v.closeOnce.Do(func (){
		
		////close all goruntines
		v.cancel()
		
		///end session
		v.Session.End()
		
		// cleanup route
    	if v.AdapterIndex != 0 {
    	    if err := v.RemoveRoute(); err != nil {
    	        fmt.Println("warning: could not remove route:", err)
    	    }
    	}

		////delete chan
		close(v.PacketChan) 

		////wait for goruntines to close
		v.wg.Wait()

		///delete interface
		v.Iface.Close()

	})

}
