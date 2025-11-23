package vna

func (v *VNA) RunReader() {
	v.wg.Add(1)
	go func() {
		defer v.wg.Done()
		for {

			if v.CtxStopped() {
				return

			}

			pkt, err := v.Session.ReceivePacket()
			if err != nil {
				//fmt.Println("read error:", err)
				continue

			}

			copyPkt := make([]byte, len(pkt))
			copy(copyPkt, pkt)

			select {
			case v.PacketChan <- copyPkt:
			default:
			}

			// free buffer floyd
			v.Session.ReleaseReceivePacket(pkt)
		}
	}()
}