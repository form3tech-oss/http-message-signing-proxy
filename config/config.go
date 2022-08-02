package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Target string       `mapstructure:"target"`
	Server ServerConfig `mapstructure:"serverConfig"`
	Signer SignerConfig `mapstructure:"signer"`
}

type ServerConfig struct {
	Port int       `mapstructure:"port"`
	TLS  TLSConfig `mapstructure:"tls"`
}

type TLSConfig struct {
	Enable       bool   `mapstructure:"enable"`
	CertFilePath string `mapstructure:"certFilePath"`
	KeyFilePath  string `mapstructure:"keyFilePath"`
}

type SignerConfig struct {
	KeyId             string   `mapstructure:"keyId"`
	KeyFilePath       string   `mapstructure:"keyFilePath"`
	BodyDigestAlgo    string   `mapstructure:"bodyDigestAlgo"`
	SignatureHashAlgo string   `mapstructure:"signatureHashAlgo"`
	SignatureHeaders  []string `mapstructure:"signatureHeaders"`
}

func GetConfig(configFilePath string) (*Config, error) {
	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
