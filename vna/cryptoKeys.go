package vna

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
)

////load crypto keys objet into vna object
func (v *VNA )LoadCryptoKeys()  error {
	
	serverPubB64 := os.Getenv("SERVER_PUBLIC_KEY")
	if serverPubB64 == "" {
		return  fmt.Errorf("SERVER_PUBLIC_KEY není nastaven v .env")
	}

	pubBytes, err := base64.StdEncoding.DecodeString(serverPubB64)
	if err != nil {
		return fmt.Errorf("SERVER_PUBLIC_KEY: špatný base64 formát: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize { 
		return fmt.Errorf("SERVER_PUBLIC_KEY: musí být %d bytů, má %d", ed25519.PublicKeySize, len(pubBytes))
	}

	cryptoKeys := CryptoKeys{
		ServerPub: ed25519.PublicKey(pubBytes), 
		SharedKey: nil,                         
	}

	v.Keys = cryptoKeys

	return nil
}

