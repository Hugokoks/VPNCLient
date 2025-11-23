package vna

import (
	"context"
	"fmt"
	"net"
	"sync"

	"golang.zx2c4.com/wintun"
)

////struct for windows
type VNA struct {
	Iface   *wintun.Adapter
	Session wintun.Session

	IfName    string
	IP 		string
	Mask    string
	AdapterIndex int
	
	ctx     context.Context
	cancel  context.CancelFunc
	closeOnce sync.Once

 	RemoteAddr  string ////Where client sending
    LocalAddr    string ///Where client listening

	wg sync.WaitGroup // producers (listeners)

	Conn *net.UDPConn
	PacketChan chan []byte
}

func New(rootCtx context.Context, ifName string,ip string,mask string) (*VNA, error) {

	///create virtual network interface
	iface, err := wintun.CreateAdapter(ifName, "Wintun", nil)

	if err != nil {
		return nil, fmt.Errorf("CreateAdapter: %w", err)
	}

	///buffer size 2 Mb
	var bufferSize uint32 = 0x200000 

	////start session
	sess, err := iface.StartSession(bufferSize)
	if err != nil {
		iface.Close()
		return nil, fmt.Errorf("StartSession: %w", err)
	}

    ctx, cancel := context.WithCancel(rootCtx)

	return &VNA{
		Iface:      iface,
		Session:    sess,
		IfName:     ifName,
		IP: 		ip,
		Mask: 		mask,
		ctx:        ctx,
		cancel:     cancel,
		PacketChan: make(chan []byte, 4096),
		RemoteAddr: "192.168.8.236:5000",
		LocalAddr: ":5000",

	}, nil
}



func (v *VNA) Start(){

	v.RunReader()
	v.RunSender()
	v.RunClientListener()

}
func (v *VNA) Stop(){

	v.Close()
}

func (v *VNA) Close() {
	v.closeOnce.Do(func (){
		
		////close all goruntines
		v.cancel()
		
		///end session
		v.Session.End()
		
		// cleanup route
    	if v.AdapterIndex != 0 {
    	    if err := v.RemoveRoute(); err != nil {
    	        fmt.Println("warning: could not remove route:", err)
    	    }
    	}


		////wait for goruntines to close
		v.wg.Wait()

		/////Close UDP connection
		v.Conn.Close()

		///delete interface
		v.Iface.Close()

	})

}
