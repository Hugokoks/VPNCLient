package vna

import (
	"fmt"
	"time"
)

func (v *VNA) RunSender() {
	v.wg.Add(1)

	go func() {
		defer v.wg.Done()

		select {
		case <-v.HandShakeDone:
		case <-v.ctx.Done():
			return
		}

		for {
			select {
			case <-v.ctx.Done():
				return

			case pkt, ok := <-v.PacketChan:
				if !ok || pkt == nil {
					return
				}

				_ = v.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

				if err := sendEncrypted(v.Aead, v.Conn, pkt); err != nil {
					fmt.Println("udp write error:", err)
				}
			}
		}
	}()
}
