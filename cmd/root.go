package cmd

import (
	"fmt"
	"os"

	"github.com/form3tech-oss/https-signing-proxy/config"
	"github.com/form3tech-oss/https-signing-proxy/proxy"
	"github.com/spf13/cobra"
)

var (
	cfg     config.Config
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flag("").Value.String()
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "")

	var setFlagValues []string
	rootCmd.Flags().StringArrayVar(&setFlagValues, "set", nil, "")

}

func NewRootCmd() *cobra.Command {
	var (
		cfgFile   string
		overrides []string
	)

	rootCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flag("").Value.String()
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
	f.StringArrayVar(&overrides, "set", nil, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")

	return rootCmd
}
