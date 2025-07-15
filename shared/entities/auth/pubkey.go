package auth

type PublicKeyEntry struct {
	KID string `yaml:"kid"`
	Key string `yaml:"key"`
}
