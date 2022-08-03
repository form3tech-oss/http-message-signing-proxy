package config

type Config struct {
	Proxy  ProxyConfig  `mapstructure:"proxy"`
	Server ServerConfig `mapstructure:"server"`
}

type ServerConfig struct {
	Port int       `mapstructure:"port"`
	TLS  TLSConfig `mapstructure:"tls"`
}

type ProxyConfig struct {
	UpstreamTarget string       `mapstructure:"upstreamTarget"`
	Signer         SignerConfig `mapstructure:"signer"`
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
