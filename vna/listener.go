package vna

import (
	"fmt"
	"time"
)

func (v *VNA) RunClientListener() {
	v.wg.Add(1)
	go func() {
		defer v.wg.Done()


		buf := make([]byte, 65535)

		for {
			if v.CtxStopped(){

				return
			}
            _ = v.Conn.SetReadDeadline(time.Now().Add(1 * time.Second)) ////wait for data max 1sec 

			n, _, err := v.Conn.ReadFromUDP(buf)
			  
			if err != nil {
                continue
            }

			pkt := make([]byte,n)
			
			copy(pkt,buf[:n])
			fmt.Printf("Listened packet, len=%d\n", len(pkt))

			v.Session.SendPacket(pkt)
		
		}
	}()
}
