package cmd

import (
	"chat-poc/internal/config"
	"os"

	"github.com/eldius/initial-config-go/setup"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/spf13/cobra"

	initCfg "github.com/eldius/initial-config-go/configs"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chat-poc",
	Short: "A simple terminal AI chat application",
	Long:  `A simple terminal AI chat application.`,
	PersistentPreRunE: setup.PersistentPreRunE(
		"my-chat-app-cli",
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
			config.BedrockChatMemoryDBPathProp,
			config.ConfluenceAuthRedirectURLProp,
			config.ConfluenceAuthURLProp,
			config.ConfluenceAuthResponseTypeProp,
			config.ConfluenceAuthAudienceProp,
			config.ConfluenceBaseURLProp,
			config.ConfluenceClientIDProp,
			config.ConfluenceScopesProp,
			config.ConfluenceAuthPromptProp,
			config.ConfluenceAuthRefreshTokenURLProp,
			config.ConfluenceClientSecretProp,
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chat-poc.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
