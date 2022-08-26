package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

var (
	kvRegex = regexp.MustCompile(`([a-zA-Z\d_.]+)=(.*)`)
)

type KV map[string]interface{}

func parseKV(ss []string) (KV, error) {
	kv := KV{}
	for _, s := range ss {
		matches := kvRegex.FindAllStringSubmatch(s, -1)
		if len(matches) == 0 {
			return nil, NewValidationError(fmt.Sprintf("invalid key-value flag format: '%s'", s))
		}
		match := matches[0]
		kv[match[1]] = match[2]
	}
	return kv, nil
}

func LoadConfig(configFilePath string, overrides []string) (*Config, error) {
	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if len(overrides) > 0 {
		kv, err := parseKV(overrides)
		if err != nil {
			return nil, err
		}
		err = viper.MergeConfigMap(kv)
		if err != nil {
			return nil, fmt.Errorf("failed to override config with flags: %w", err)
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
