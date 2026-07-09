package llm

import "github.com/spf13/viper"

const (
	ollamaConfigKey = "ollama"
)

// Opts is the configuration for the Ollama backend
type Opts struct {
	Endpoint   string         `yaml:"endpoint" json:"endpoint"`
	Generation GenerationOpts `yaml:"generation" json:"generation"`
}

type GenerationOpts struct {
	Model string `yaml:"model" json:"model"`
	Cache struct {
		Enabled   bool    `yaml:"enabled" json:"enabled"`
		Threshold float64 `yaml:"threshold" json:"threshold"`
	} `yaml:"cache" json:"cache"`
	Context struct {
		MaxDocuments       int64   `yaml:"max_documents" json:"max_documents"`
		MinSimilarityScore float64 `yaml:"min_similarity_score" json:"min_similarity_score"`
		Enabled            bool    `yaml:"enabled" json:"enabled"`
	} `yaml:"context" json:"context"`
}

func LoadOllamaOpts() (Opts, error) {
	var opts Opts
	if err := viper.UnmarshalKey(ollamaConfigKey, &opts); err != nil {
		return opts, err
	}
	return opts, nil
}
