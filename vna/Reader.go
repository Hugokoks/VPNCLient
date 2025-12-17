package vna

import "time"

func (v *VNA) runReader() {

	defer v.wg.Done()

	for {
		if v.ctxStopped() {
			return
		}
		////win tun pkt
		pkt, err := v.Session.ReceivePacket()

		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		///create own pkt
		copyPkt := make([]byte, len(pkt))
		copy(copyPkt, pkt)

		select {
		case v.PacketChan <- copyPkt:
		default:

		}

		// free buffer floyd
		v.Session.ReleaseReceivePacket(pkt)
	}
}
