package logger

import (
	"testing"

	"github.com/form3tech-oss/http-message-signing-proxy/config"
	"github.com/stretchr/testify/require"
)

type errCheckFn func(require.TestingT, error, ...interface{})

func TestConfigure(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.LogConfig
		errCheckFn errCheckFn
	}{
		{
			"valid config",
			config.LogConfig{
				Level:  "info",
				Format: "json",
			},
			require.NoError,
		},
		{
			"invalid level",
			config.LogConfig{
				Level:  "oompa",
				Format: "json",
			},
			require.Error,
		},
		{
			"invalid format",
			config.LogConfig{
				Level:  "info",
				Format: "xml",
			},
			require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Configure(test.cfg)
			test.errCheckFn(t, err)
		})
	}
}
