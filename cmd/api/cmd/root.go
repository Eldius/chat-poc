package cmd

import (
	"chat-poc/internal/config"
	"os"

	initCfg "github.com/eldius/initial-config-go/configs"
	"github.com/eldius/initial-config-go/setup"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chat-poc",
	Short: "A simple API to be used to find transaction status",
	Long:  `A simple API to be used to find transaction status.`,
	PersistentPreRunE: setup.PersistentPreRunE(
		"my-chat-app-api",
		setup.WithConfigFileToBeUsed(cfgFile),
		setup.WithDefaultCfgFileLocations("."),
		setup.WithDefaultCfgFileName("config"),
		setup.WithEnvPrefix("CHAT"),
		setup.WithDefaultValues(map[string]any{
			initCfg.LogLevelKey:      initCfg.LogLevelDEBUG,
			initCfg.LogFormatKey:     initCfg.LogFormatJSON,
			initCfg.LogOutputFileKey: "execution.log",
		}),
		setup.WithProps(
			config.BedrockRegionProp,
			config.BedrockInferenceModelProp,
			config.BedrockEmbeddingModelProp,
			config.BedrockInferenceTemperatureProp,
			config.BedrockInferenceMaxIterationsProp,
			config.BedrockInferenceTopKProp,
			config.BedrockInferenceTopPProp,
			config.BedrockCacheEnabledProp,
			config.BedrockCachePersistTimeoutProp,
			config.BedrockCacheDBPathProp,
			config.APIPortProp,
			config.BedrockChatMemoryDBPathProp,
		),
		setup.WithOpenTelemetryOptions(
			telemetry.WithOtelEnabled(false),
			telemetry.WithService(config.AppName, config.Version, "local"),
		),
	),
}

var (
	cfgFile string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chat-poc.yaml)")
}
