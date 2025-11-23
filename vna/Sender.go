package vna

import (
	"fmt"
	"time"
)

func (v *VNA) RunSender() {
    v.wg.Add(1)

    go func() {
        defer v.wg.Done()

        for {
            
        
            select{
            
            case <- v.ctx.Done():
                return
            
            case pkt, ok := <-v.PacketChan:
                
                if !ok || pkt == nil {
                    return 
                }
              
                v.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second)) 
                _, err := v.Conn.Write(pkt)
                
                if err != nil {

                    fmt.Println("udp write error:", err)
                }
            }  
        
        }
        
    }()
}