package base64

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
)

func MarshalPubKey(pub ed25519.PublicKey) string {
	if pub == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(pub)
}

func UnmarshalPubKey(s string) (ed25519.PublicKey, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: %d", len(b))
	}
	return b, nil
}
