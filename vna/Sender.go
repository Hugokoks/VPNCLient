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

			payload := make([]byte, 0, len(v.ClientID)+len(encrypted))
			payload = append(payload, v.ClientID...)
			payload = append(payload, encrypted...)

			typedPkt := buildPacket(PacketData,payload)
			
			_ = v.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

			if _,err := v.Conn.Write(typedPkt); err != nil{
				log.Println("udp write error:", err)
			}

		}
	}
}
