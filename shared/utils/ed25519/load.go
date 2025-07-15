package ed25519

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/kxddry/lectura/shared/entities/auth"
	"github.com/kxddry/lectura/shared/utils/base64"
	"os"
	"strings"
)

func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("invalid private key PEM")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	edPriv, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("Not an Ed25519 Private Key")
	}
	return edPriv, nil
}

func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	edPub, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("Not an Ed25519 Public Key")
	}
	return edPub, nil
}

func LoadPublicKeys(entries []auth.PublicKeyEntry) (map[string]ed25519.PublicKey, error) {
	result := make(map[string]ed25519.PublicKey)
	for _, entry := range entries {
		var keyBytes []byte
		var err error

		if strings.HasPrefix(entry.Key, "./") || strings.HasSuffix(entry.Key, ".pub") {
			keyBytes, err = os.ReadFile(entry.Key)
			if err != nil {
				return nil, err
			}
			keyBytes = bytes.TrimSpace(keyBytes)
		} else {
			keyBytes, err = base64.UnmarshalPubKey(entry.Key)
			if err != nil {
				return nil, errors.New("invalid base64 public key for kid " + entry.KID)
			}
		}

		if len(keyBytes) != ed25519.PublicKeySize {
			return nil, errors.New("invalid pubkey size for kid " + entry.KID)
		}
		result[entry.KID] = ed25519.PublicKey(keyBytes)
	}
	return result, nil
}
