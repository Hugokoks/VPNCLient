package vna

import (
	"fmt"
	"net"
	"time"
)

func (v *VNA) RequestIP()  error {

	// Send request
	//// req packet [0] = PacketIPResponse
	req := []byte{byte(PacketIPRequest)}

	_ = v.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := v.Conn.Write(req); err != nil {
		return fmt.Errorf("IP request send failed: %w", err)
	}

	// Read response
	/*
				///9 bytes
		    	[0] = PacketIPResponse
		    	[1] = IP oktet 1
		    	[2] = IP oktet 2
		    	[3] = IP oktet 3
		    	[4] = IP oktet 4
		    	[5] = MASK oktet 1
		    	[6] = MASK oktet 2
		    	[7] = MASK oktet 3
		    	[8] = MASK oktet 4
	*/

	buf := make([]byte, 9)

	_ = v.Conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := v.Conn.Read(buf)

	if err != nil {
		return  fmt.Errorf("IP request read failed: %w", err)
	}

	if n < 9 {
		return  fmt.Errorf("invalid IP response length: %d", n)
	}

	// Parse
	if PacketType(buf[0]) != PacketIPResponse {
		return fmt.Errorf("unexpected packet type: %d", buf[0])
	}

	ip := net.IP(buf[1:5]).String()
	mask := net.IP(buf[5:9]).String()

	v.IP = ip
	v.Mask = mask
	return nil
}