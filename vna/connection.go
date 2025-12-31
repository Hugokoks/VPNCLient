package vna

import "net"

///create UDP connection with server
func (v *VNA) InitConnection() error {

	///Client Address
	laddr, err := net.ResolveUDPAddr("udp", v.LocalAddr)
	if err != nil {
		return err
	}

	///Server Address
	raddr, err := net.ResolveUDPAddr("udp", v.RemoteAddr)
	if err != nil {
		return err
	}

	////UDP Connection between Client - Server
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}

	v.Conn = conn
	return nil

}