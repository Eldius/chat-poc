package ollama

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadOllamaOpts(t *testing.T) {
	viper.Set("ollama.endpoint", "http://test:11434")
	viper.Set("ollama.generation.model", "test-model")
	t.Cleanup(func() {
		viper.Reset()
	})

	opts, err := LoadOllamaOpts()
	assert.NoError(t, err)
	assert.Equal(t, "http://test:11434", opts.Endpoint)
	assert.Equal(t, "test-model", opts.Generation.Model)
}

func TestLoadOllamaOptsDefaults(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	opts, err := LoadOllamaOpts()
	assert.NoError(t, err)
	assert.Empty(t, opts.Endpoint)
	assert.Empty(t, opts.Generation.Model)
}
