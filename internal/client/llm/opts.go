package llm

import (
	"encoding/json"
	"github.com/spf13/viper"
	"log/slog"
)

const (
	backendConfigKey = "backend"
)

// Opts is the configuration for the Ollama backend
type Opts struct {
	Endpoint   string         `yaml:"endpoint" json:"endpoint"`
	Type       BackendType    `yaml:"type" json:"type"`
	Generation GenerationOpts `yaml:"generation" json:"generation"`
	Key        string         `yaml:"api_key" json:"api_key"`
}

type GenerationOpts struct {
	Model   string  `yaml:"model" json:"model"`
	Think   bool    `yaml:"think" json:"think"`
	Cache   Cache   `yaml:"cache" json:"cache"`
	Context Context `yaml:"context" json:"context"`
}

type Cache struct {
	Enabled   bool    `yaml:"enabled" json:"enabled"`
	Threshold float64 `yaml:"threshold" json:"threshold"`
}

type Context struct {
	MaxDocuments       int64   `yaml:"max_documents" json:"max_documents"`
	MinSimilarityScore float64 `yaml:"min_similarity_score" json:"min_similarity_score"`
	Enabled            bool    `yaml:"enabled" json:"enabled"`
}

func LoadOpts() (Opts, error) {
	var opts Opts
	if err := viper.UnmarshalKey(backendConfigKey, &opts); err != nil {
		return opts, err
	}

	//if b, err := yaml.Marshal(opts); err == nil {
	if b, err := json.MarshalIndent(opts, "", "    "); err == nil {
		slog.With("backend_opts", string(b)).Info("backend opts")
	}

	return opts, nil
}
