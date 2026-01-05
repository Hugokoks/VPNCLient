package keys

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// LoadOrCreateClientIdentity loads an existing client identity key
// or generates and persists a new one on first run.
func (c * CryptoKeys)LoadOrCreateClientIdentity()  error {

	home, err := os.UserHomeDir()
	if err != nil {
		return  err
	}

	dir := filepath.Join(home, ".myvpn")
	keyPath := filepath.Join(dir, "client.key")

	////try to create dir 
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Load existing key
	if data, err := os.ReadFile(keyPath); err == nil {
		
		privBytes, err := base64.StdEncoding.DecodeString(string(data))
		
		if err != nil {
			return fmt.Errorf("invalid client key format: %w", err)
		}
		
		///LEN GUARD
		if len(privBytes) != ed25519.PrivateKeySize {
    		return fmt.Errorf("invalid private key size")
		}

		////convert into ed25519.PrivateKey object
		priv := ed25519.PrivateKey(privBytes)
		////Count public key 
		pub := priv.Public().(ed25519.PublicKey)
		
		c.ClientPriv = priv
		c.ClientPub = pub
		
		return nil
	}

	// Generate new identity keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	privB64 := base64.StdEncoding.EncodeToString(priv)
	if err := os.WriteFile(keyPath, []byte(privB64), 0600); err != nil {
		return  err
	}

	c.ClientPriv = priv
	c.ClientPub = pub

	return  nil
}
