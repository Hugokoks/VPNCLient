package keys

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
)

// LoadServerPublicKey loads the trusted server public key from env.
func (c *CryptoKeys)LoadServerPublicKey() error {

	serverPubB64 := os.Getenv("SERVER_PUBLIC_KEY")
	if serverPubB64 == "" {
		return  fmt.Errorf("SERVER_PUBLIC_KEY not set")
	}

	pubBytes, err := base64.StdEncoding.DecodeString(serverPubB64)
	if err != nil {
		return fmt.Errorf("SERVER_PUBLIC_KEY invalid base64: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize {
		return fmt.Errorf(
			"SERVER_PUBLIC_KEY must be %d bytes, got %d",
			ed25519.PublicKeySize,
			len(pubBytes),
		)
	}

	c.ServerPub = ed25519.PublicKey(pubBytes)
	return nil
}
