package vna

import (
	"fmt"
)

func (v *VNA) RunClientListener() {
    v.wg.Add(1)
    go func() {
        defer v.wg.Done()

        buf := make([]byte, 65535)

        for {
            if v.CtxStopped() {
                return
            }

            ipPkt, err := recvDecrypted(v.Aead, v.Conn,buf)
            if err != nil {
                continue
            }
            sendBuf, err := v.Session.AllocateSendPacket(len(ipPkt))
            if err != nil {
                fmt.Println("AllocateSendPacket error:", err)
                continue
            }
            copy(sendBuf, ipPkt)

            fmt.Printf("Listened packet, len=%d\n", len(ipPkt))
            v.Session.SendPacket(sendBuf)
        }
    }()
}