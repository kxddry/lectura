package s3

type StorageConfig struct {
	Endpoint  string `yaml:"endpoint" env-required:"true"`
	PublicURL string `yaml:"public_url" env-required:"true"`
	AccessKey string `yaml:"access_key" env-required:"true"`
	Secret    string `yaml:"secret" env-required:"true"`
	UseSSL    bool   `yaml:"ssl" env-default:"false"`
}
