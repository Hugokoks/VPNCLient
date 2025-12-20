package vna

import (
	"VPNClient/crypted"
	"fmt"
	"log"
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
						
			encrypted ,err := crypted.SendEncrypted(v.Aead, pkt); 

			if err != nil {
				fmt.Println("udp write error:", err)
			}

			typedPkt := buildPacket(PacketData,encrypted)
			
			_ = v.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

			if _,err := v.Conn.Write(typedPkt); err != nil{
				log.Println("udp write error:", err)
			}

		}
	}
}
