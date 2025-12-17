package vna

import (
	"VPNClient/crypted"
	"fmt"
	"time"
)

func (v *VNA) runClientListener() {
    defer v.wg.Done()

    buf := make([]byte, 65535)

    for {
        if v.ctxStopped() {
            return
        }
        v.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))

        ipPkt, err := crypted.RecvDecrypted(v.Aead, v.Conn,buf)
        
        if err != nil {
            time.Sleep(10 * time.Millisecond)
            continue
        }

        ////AllocatePacket for network driver
        sendBuf, err := v.Session.AllocateSendPacket(len(ipPkt))
        
        if err != nil {
            fmt.Println("AllocateSendPacket error:", err)
            continue
        }
        copy(sendBuf, ipPkt)

        fmt.Printf("Listened packet, len=%d\n", len(ipPkt))
        
        ////write packet in network card
        v.Session.SendPacket(sendBuf)
    }
}