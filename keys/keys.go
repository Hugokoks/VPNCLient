package keys

import "crypto/ed25519"

// CryptoKeys holds all cryptographic material needed by the client.

type CryptoKeys struct {
	// Server identity (trusted)
	ServerPub ed25519.PublicKey

	// Client identity (long-term)
	ClientPriv ed25519.PrivateKey
	ClientPub  ed25519.PublicKey

	// Session key (ephemeral, set after handshake)
	SharedKey []byte
}