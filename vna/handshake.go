package vna

import (
	"VPNClient/crypted"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/curve25519"
)

// Handshake performs an authenticated key exchange with the server.
//
// The protocol flow is:
//  1. Client generates an ephemeral Curve25519 key pair.
//  2. Client sends its ephemeral public key to the server.
//  3. Server responds with its own ephemeral public key,
//     signed by the server's long-term Ed25519 private key.
//  4. Client verifies the server signature (authentication).
//  5. Both sides derive a shared secret using ECDH.
//  6. A symmetric AEAD key is derived from the shared secret.
//
// If any step fails, the handshake is aborted and no crypto state is installed.

func (v *VNA) Handshake() error {

	// ---------------------------------------------------------------------
	// Generate client's ephemeral private key
	// ---------------------------------------------------------------------
	var clientEphPriv [32]byte
	if _, err := io.ReadFull(rand.Reader, clientEphPriv[:]); err != nil {
		return fmt.Errorf("generování efemerního klíče: %w", err)
	}

	// Clamp the private key as required by Curve25519.	
	clientEphPriv[0] &= 248
	clientEphPriv[31] &= 127
	clientEphPriv[31] |= 64

	// ---------------------------------------------------------------------
	// 2. Compute client's ephemeral public key
	// ---------------------------------------------------------------------	
	var clientEphPub [32]byte
	curve25519.ScalarBaseMult(&clientEphPub, &clientEphPriv)

	// ---------------------------------------------------------------------
	// Send client's ephemeral public key to the server
	// --------------------------------------------------------------------

	signatureReq := ed25519.Sign(v.Keys.ClientPriv, clientEphPub[:])

	payload := make([]byte, 0, 32+32+64)
	payload = append(payload, v.Keys.ClientPub...)   // client identity
	payload = append(payload, clientEphPub[:]...)    // ephemeral pub
	payload = append(payload, signatureReq...)   


	pkt := buildPacket(PacketHandshakeReq,payload)

	if _, err := v.Conn.Write(pkt); err != nil {
		return fmt.Errorf("failed to send client pub: %w", err)
	}

	// ---------------------------------------------------------------------
	//  Receive server response:
	//    [32 bytes server ephemeral public key]
	//    [64 bytes Ed25519 signature over server ephemeral public key]
	// ---------------------------------------------------------------------	
	res := make([]byte, 1+32+32+64)
	if _, err := io.ReadFull(v.Conn, res); err != nil {
		return fmt.Errorf("failed to read server response: %w", err)
	}

	if PacketType(res[0]) != PacketHandshakeRes{

		return fmt.Errorf("unexpected packet type: %d", res[0])
	}

	payloadRes := res[1:]

	clientID     := payloadRes[0:32]
	serverEphPub := payloadRes[32:64]
	signatureRes := payloadRes[64:128]

	// ---------------------------------------------------------------------
	// Verify server identity
	// ---------------------------------------------------------------------
	// The server signs its ephemeral public key with its long-term Ed25519 key.
	// This prevents MITM attacks and authenticates the server.
	signedData := make([]byte, 0, 64)

	signedData = append(signedData, clientID...)
	signedData = append(signedData, serverEphPub...)

	if !ed25519.Verify(v.Keys.ServerPub, signedData, signatureRes) {
		return fmt.Errorf("NEPLATNÝ PODPIS SERVERU – možné MITM")
	}

	// ---------------------------------------------------------------------
	// 6. Perform ECDH to derive the shared secret
	// ---------------------------------------------------------------------
	var serverPub [32]byte
	copy(serverPub[:], serverEphPub)
	
	
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, &clientEphPriv, &serverPub)

	// ---------------------------------------------------------------------
	// 7. Derive symmetric encryption key from shared secret
	// ---------------------------------------------------------------------
	// The shared secret is hashed to produce a fixed-length AEAD key.
	h := sha256.Sum256(sharedSecret[:])
	v.Keys.SharedKey = h[:]
	
	fmt.Printf("Shared key: %x", v.Keys.SharedKey)
	
	// Initialize AEAD cipher used for encrypted transport.
	aead, err := crypted.NewAEAD(v.Keys.SharedKey) 
	if err != nil {
	    return err
	}
	v.Aead = aead
	
	// Initialize Client ID
	v.ClientID = make([]byte, 32)
	copy(v.ClientID, clientID)


	fmt.Println("Handshake úspěšný – šifrovací klíč připraven")
	return nil
}

////attempt handshake multiple times
func (v *VNA) handshakeLoop(maxRetries int){

	defer v.wg.Done()

	for i := 1; i <= maxRetries; i++ {
		fmt.Printf("Handshake attempt %d/%d\n", i, maxRetries)

		if err := v.Handshake(); err == nil {
			close(v.handshakeReady) 
			fmt.Println("Handshake successful")
			return
		}

		select {
		
		///try again delay	
		case <-time.After(2 * time.Second):
		
		case <-v.ctx.Done():
			return
		}
	}

	fmt.Println("Handshake failed after retries")
	v.cancel()
}

/////for nonblocking handshake start
func (v *VNA) isHandshakeReady() bool {
	select {
	case <-v.handshakeReady:
		
		return true
	
	default:
		return false
	}
}