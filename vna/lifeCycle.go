package vna

import "fmt"

func (v *VNA) Start() {
	v.wg.Add(1)
	go v.handshakeLoop(3)

	select {
	case <-v.handshakeReady:
		// handshake OK, pokračujeme
	case <-v.ctx.Done():
		// shutdown request (Ctrl+C)
		return
	}

	v.wg.Add(1)
	go v.runReader()

	v.wg.Add(1)
	go v.runSender()

	v.wg.Add(1)
	go v.runClientListener()
}

func (v *VNA) Stop() {

	v.Close()
}

func (v *VNA) Close() {
	v.closeOnce.Do(func() {

		// Signal all goroutines to stop
		v.cancel()

		//Wait until ALL goroutines exit
		// (no one is using Session / Conn anymore)
		v.wg.Wait()

		// End Wintun session (no goroutine is touching it now)
		v.Session.End()

		// Close UDP socket
		if v.Conn != nil {
			_ = v.Conn.Close()
		}

		// Remove route
		if v.AdapterIndex != 0 {
			if err := v.RemoveRoute(); err != nil {
				fmt.Println("warning: could not remove route:", err)
			}
		}

		// Close Wintun adapter
		if v.Iface != nil {
			v.Iface.Close()
		}
	})
}