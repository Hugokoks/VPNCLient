package vna

import (
	"VPNClient/keys"
	"context"
	"crypto/cipher"
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
	handshakeReady chan struct{}


	////In future make client struct
	Aead cipher.AEAD
	Keys keys.CryptoKeys
	ClientID []byte

}

func New(rootCtx context.Context, ifName string,remoteAddr string,localAddr string) (*VNA, error) {

	///create virtual network interface
	iface, err := wintun.CreateAdapter(ifName, "Wintun", nil)

	if err != nil {
		return nil, fmt.Errorf("CreateAdapter: %w", err)
	}

	///buffer size 8 Mb
	var bufferSize uint32 = 0x800000 

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
		ctx:        ctx,
		cancel:     cancel,
		PacketChan: make(chan []byte, 16384),
		RemoteAddr: remoteAddr,
		LocalAddr: localAddr ,
		handshakeReady: make(chan struct{}),

	}, nil
	
}

