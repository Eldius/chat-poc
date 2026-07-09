package llm

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadOllamaOpts(t *testing.T) {
	viper.Set("backend.endpoint", "http://test:11434")
	viper.Set("backend.generation.model", "test-model")
	t.Cleanup(func() {
		viper.Reset()
	})

	opts, err := LoadOpts()
	assert.NoError(t, err)
	assert.Equal(t, "http://test:11434", opts.Endpoint)
	assert.Equal(t, "test-model", opts.Generation.Model)
}

func TestLoadOllamaOptsDefaults(t *testing.T) {
	t.Cleanup(func() {
		viper.Reset()
	})

	opts, err := LoadOpts()
	assert.NoError(t, err)
	assert.Empty(t, opts.Endpoint)
	assert.Empty(t, opts.Generation.Model)
}
