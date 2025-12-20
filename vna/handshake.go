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
	// Ensure the server's long-term public key is loaded.
	// Without this key, the server identity cannot be verified.
	if len(v.Keys.ServerPub) == 0 {
		return fmt.Errorf("serverový veřejný klíč není načten")
	}

	// ---------------------------------------------------------------------
	// Generate client's ephemeral private key
	// ---------------------------------------------------------------------
	var clientPriv [32]byte
	if _, err := io.ReadFull(rand.Reader, clientPriv[:]); err != nil {
		return fmt.Errorf("generování efemerního klíče: %w", err)
	}

	// Clamp the private key as required by Curve25519.	
	clientPriv[0] &= 248
	clientPriv[31] &= 127
	clientPriv[31] |= 64

	// ---------------------------------------------------------------------
	// 2. Compute client's ephemeral public key
	// ---------------------------------------------------------------------	
	var clientPub [32]byte
	curve25519.ScalarBaseMult(&clientPub, &clientPriv)

	// ---------------------------------------------------------------------
	// Send client's ephemeral public key to the server
	// --------------------------------------------------------------------

	pkt := buildPacket(PacketHandshake,clientPub[:])

	if _, err := v.Conn.Write(pkt); err != nil {
		return fmt.Errorf("selhalo odeslání client pub: %w", err)
	}

	// ---------------------------------------------------------------------
	//  Receive server response:
	//    [32 bytes server ephemeral public key]
	//    [64 bytes Ed25519 signature over server ephemeral public key]
	// ---------------------------------------------------------------------	
	response := make([]byte, 96)
	if _, err := io.ReadFull(v.Conn, response); err != nil {
		return fmt.Errorf("selhalo čtení odpovědi serveru: %w", err)
	}

	serverEphPub := response[:32]
	signature := response[32:]

	// ---------------------------------------------------------------------
	// Verify server identity
	// ---------------------------------------------------------------------
	// The server signs its ephemeral public key with its long-term Ed25519 key.
	// This prevents MITM attacks and authenticates the server.
	if !ed25519.Verify(v.Keys.ServerPub, serverEphPub, signature) {
		return fmt.Errorf("NEPLATNÝ PODPIS SERVERU – možné MITM!")
	}

	// ---------------------------------------------------------------------
	// 6. Perform ECDH to derive the shared secret
	// ---------------------------------------------------------------------
	var serverPub [32]byte
	copy(serverPub[:], serverEphPub)
	
	
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, &clientPriv, &serverPub)

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