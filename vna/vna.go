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
	Session  wintun.Session
	name    string
	ctx     context.Context
	cancel  context.CancelFunc
	
	wg     sync.WaitGroup

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
		Iface: iface, 
		Session: sess, 
		name: name, 
		ctx: ctx, 
		cancel: cancel,
		}, nil
}

// RunListener spustí goroutine, která čte pakety a předává je do handleru.
// handler musí do sebe kopírovat data pokud je bude chtít zpracovat asynchronně.
func (v *VNA) RunListener(handler func([]byte)) {
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
				// když session.End() zavolá, ReceivePacket obvykle vrátí chybu -> ukonči smyčku
				// pokud chceš ignorovat krátkodobé chyby, můžeš je retryovat s pauzou
				select {
				case <-v.ctx.Done():
					return
				default:
				}
				// logni a čekej malinko před dalším pokusem
				log.Printf("ReceivePacket error: %v; retrying...", err)
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// předání paketu handleru (pokud potřebuješ aby handler neběhal v této goroutine,
			// může handler sám poslat kopii paketu do kanálu/gorutiny)
			handler(packet)

			// uvolnit buffer
			v.Session.ReleaseReceivePacket(packet)
		}
	}()
}

// Close bezpečně ukončí listener a uvolní resources
func (v *VNA) Close() {
	// 1) signalizuj gorutinám, aby se ukončily
	v.cancel()

	// 2) ukonči session -- tím se uvolní blokované ReceivePacket volání
	//    (kdybys čekal jen na cancel, ReceivePacket může zůstat blokované)
	v.Session.End()

	// 3) počkej na ukončení gorutin
	v.wg.Wait()

	// 4) uzavři adapter
	v.Iface.Close()
}