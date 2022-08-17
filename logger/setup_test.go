package logger

import (
	"testing"

	"github.com/form3tech-oss/http-message-signing-proxy/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type errCheckFn func(require.TestingT, error, ...interface{})

func TestConfigure(t *testing.T) {
	tests := []struct {
		name              string
		cfg               config.LogConfig
		errCheckFn        errCheckFn
		expectedFormatter log.Formatter
		expectedLevel     log.Level
	}{
		{
			"valid config",
			config.LogConfig{
				Level:  "info",
				Format: "json",
			},
			require.NoError,
			&log.JSONFormatter{},
			log.InfoLevel,
		},
		{
			"invalid level",
			config.LogConfig{
				Level:  "oompa",
				Format: "json",
			},
			require.Error,
			nil,
			0,
		},
		{
			"invalid format",
			config.LogConfig{
				Level:  "info",
				Format: "xml",
			},
			require.Error,
			nil,
			0,
		},
		{
			"default level and format",
			config.LogConfig{
				Level:  "",
				Format: "",
			},
			require.NoError,
			&log.TextFormatter{},
			log.InfoLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Configure(test.cfg)
			test.errCheckFn(t, err)
			if err == nil {
				require.IsType(t, test.expectedFormatter, log.StandardLogger().Formatter)
				require.Equal(t, test.expectedLevel, log.GetLevel())
			}
		})
	}
}
