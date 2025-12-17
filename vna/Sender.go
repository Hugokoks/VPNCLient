package vna

import (
	"VPNClient/crypted"
	"fmt"
	"time"
)

func (v *VNA) runSender() {

	defer v.wg.Done()

	for {
		select {
		case <-v.ctx.Done():
			return
			
		case pkt, ok := <-v.PacketChan:
			if !ok || pkt == nil {
				return
			}
			_ = v.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
			if err := crypted.SendEncrypted(v.Aead, v.Conn, pkt); err != nil {
				fmt.Println("udp write error:", err)
			}
		}
	}
}
