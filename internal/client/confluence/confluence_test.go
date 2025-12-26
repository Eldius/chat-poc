package confluence

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLParse(t *testing.T) {
	u, err := url.Parse("http://localhost:9999/auth/result")
	assert.NoError(t, err)
	assert.Equal(t, "http", u.Scheme)
	assert.Equal(t, "localhost:9999", u.Host)
	assert.Equal(t, "/auth/result", u.Path)
	assert.Equal(t, "9999", u.Port())
	t.Logf("port: %s", u.Port())
}
