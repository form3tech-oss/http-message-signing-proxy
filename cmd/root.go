package cmd

import (
	"fmt"

	"github.com/form3tech-oss/https-signing-proxy/config"
	"github.com/form3tech-oss/https-signing-proxy/logger"
	"github.com/form3tech-oss/https-signing-proxy/proxy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Execute() {
	rootCmd := NewRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		logrus.WithError(err).Fatal("failed to start server")
	}
}

func NewRootCmd() *cobra.Command {
	var (
		cfgFile   string
		overrides []string
	)

	rootCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cfgFile, overrides)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			err = logger.Configure(cfg.Log)
			if err != nil {
				return fmt.Errorf("failed to configure logger: %w", err)
			}

			signingProxy, err := proxy.NewProxy(cfg.Proxy)
			if err != nil {
				return fmt.Errorf("failed to create signing proxy: %w", err)
			}
			handler := proxy.NewHandler(signingProxy)
			server := proxy.NewServer(cfg.Server, handler)
			server.Start()

			return nil
		},
	}

	f := rootCmd.Flags()
	f.StringVar(&cfgFile, "config", "", "path to config file")
	f.StringArrayVar(&overrides, "set", nil, "set value for certain config fields to override config file, can be set multiple times")

	return rootCmd
}
