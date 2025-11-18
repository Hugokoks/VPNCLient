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
	name    string
	ctx     context.Context
	cancel  context.CancelFunc

	wg sync.WaitGroup // producers (listeners)

	PacketChan chan []byte
}

func New(name string, bufferSize uint32) (*VNA, error) {

	///create virtual network interface
	iface, err := wintun.CreateAdapter(name, "Wintun", nil)

	if err != nil {
		return nil, fmt.Errorf("CreateAdapter: %w", err)
	}

	////start session
	sess, err := iface.StartSession(bufferSize)
	if err != nil {
		iface.Close()
		return nil, fmt.Errorf("StartSession: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &VNA{
		Iface:      iface,
		Session:    sess,
		name:       name,
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

		// 🛑 TATO SMYČKA ČTE DATA A TÍM UVOLŇUJE KANÁL
		for rawPacket := range v.PacketChan {

			// Logika detekce ICMP zůstává stejná, ale nyní se provede
			if len(rawPacket) >= 20 && rawPacket[9] == 1 {
				log.Println("ICMP paket přijat a zpracován.")
			}

			// Zde by normálně probíhalo šifrování a odeslání do sítě.
			// time.Sleep(1 * time.Microsecond) // Můžete přidat pro simulaci zátěže

			packetCount++
			if time.Since(lastLog) >= 1*time.Second {
				log.Printf("[Encryptor] Rychlost: %d paketů/s", packetCount)
				packetCount = 0
				lastLog = time.Now()
			}
		}
		log.Println("[Encryptor] Ukončeno zpracování INBOUND paketů.")
	}()
}

// Close bezpečně ukončí listener a uvolní resources
func (v *VNA) Close() {
	// 1) signalizuj gorutinám, aby se ukončily
	v.cancel()

	// 2) ukonči session -- tím se uvolní blokované ReceivePacket volání
	//    (kdybys čekal jen na cancel, ReceivePacket může zůstat blokované)
	v.Session.End()

	close(v.PacketChan) ////zavreme kanal
	// 3) počkej na ukončení gorutin
	v.wg.Wait()

	// 4) uzavři adapter
	v.Iface.Close()
}
