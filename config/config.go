package config

type Config struct {
	Proxy  ProxyConfig  `mapstructure:"proxy"`
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port int       `mapstructure:"port"`
	SSL  SSLConfig `mapstructure:"ssl"`
}

type ProxyConfig struct {
	UpstreamTarget string       `mapstructure:"upstreamTarget"`
	Signer         SignerConfig `mapstructure:"signer"`
}

type SSLConfig struct {
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

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}
