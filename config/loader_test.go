package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseKV(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		out  KV
		err  error
	}{
		{
			"single vanilla",
			[]string{
				"key=value",
			},
			KV{
				"key": "value",
			},
			nil,
		},
		{
			"multiple vanilla",
			[]string{
				"key1=value1",
				"key2=value2",
			},
			KV{
				"key1": "value1",
				"key2": "value2",
			},
			nil,
		},
		{
			"keys with allowed characters",
			[]string{
				"key1.18_abcd=value",
				"._12ab=value",
			},
			KV{
				"key1.18_abcd": "value",
				"._12ab":       "value",
			},
			nil,
		},
		{
			"values with special characters",
			[]string{
				"key1=http://localhost:8000/q?sth=sth",
				"key2=v=[]{}()<>.';\"!@£$%^&*_+-",
			},
			KV{
				"key1": "http://localhost:8000/q?sth=sth",
				"key2": "v=[]{}()<>.';\"!@£$%^&*_+-",
			},
			nil,
		},
		{
			"empty value",
			[]string{
				"key=",
			},
			KV{
				"key": "",
			},
			nil,
		},
		{
			"empty key is invalid",
			[]string{
				"=value",
			},
			nil,
			&ValidationError{},
		},
		{
			"single invalid separator",
			[]string{
				"key?value",
			},
			nil,
			&ValidationError{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := parseKV(test.in)
			if test.err != nil {
				require.ErrorAs(t, err, &test.err)
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, test.out, actual)
		})
	}
}

func TestLoadConfig(t *testing.T) {

	expectedCfg := Config{
		Proxy: ProxyConfig{
			UpstreamTarget: "https://princesscarolyn.net",
			Signer: SignerConfig{
				KeyId:             "6f33b219-137c-467e-9a61-f61040a03363",
				KeyFilePath:       "/etc/form3/private/private.key",
				BodyDigestAlgo:    "SHA-512",
				SignatureHashAlgo: "SHA-256",
				SignatureHeaders: []string{
					"(request-target)",
					"host",
					"date",
				},
			},
		},
		Server: ServerConfig{
			Port: 9090,
			TLS: TLSConfig{
				Enable:       true,
				CertFilePath: "/etc/ssl/certs/cert.crt",
				KeyFilePath:  "/etc/ssl/private/private.key",
			},
		},
	}

	configFile := "config_test.yaml"
	overrides := []string{
		"proxy.upstreamTarget=https://princesscarolyn.net",
		"log.level=debug",
	}
	_ = os.Setenv("SERVER_PORT", "9090")
	_ = os.Setenv("PROXY_SIGNER_BODYDIGESTALGO", "SHA-512")

	cfg, err := LoadConfig(configFile, overrides)
	require.Nil(t, err)
	require.Equal(t, expectedCfg, *cfg)
}
