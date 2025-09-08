package chat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	smallTestString          = "test string"
	bigTestString            = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed sagittis id massa id maximus. Maecenas a nunc vitae enim dictum mollis placerat quis erat. Vestibulum consequat felis quis ex dignissim, non pharetra arcu sollicitudin. Mauris interdum arcu quis lorem fringilla, vitae scelerisque lacus sagittis. Duis eget sem a felis luctus interdum non in turpis. Nunc dapibus fringilla dui, sit amet sodales massa. In sit amet dolor tincidunt, eleifend augue et, consequat orci. Aenean porta mollis ipsum a laoreet. In ultricies pharetra tortor non ultrices. Donec mi est, maximus et nulla sit amet, fermentum imperdiet mauris. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. In justo purus, vestibulum id massa ac, lacinia porttitor sapien. Praesent commodo convallis neque, eget imperdiet metus varius ac."
	simpleTestString         = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	simpleTestStringWrapperd = "Lorem ipsum\ndolor sit\namet,\nconsectetur\nadipiscing\nelit."
)

func TestWordWrap(t *testing.T) {
	t.Run("small string with no need to line break", func(t *testing.T) {
		wrapped := WordWrap(smallTestString, 12)
		assert.Equal(t, smallTestString, wrapped)
	})

	t.Run("small string with the need to line break", func(t *testing.T) {
		wrapped := WordWrap(smallTestString, 10)
		t.Logf("wrapped: '%s'", wrapped)
		assert.Equal(t, "test\nstring", wrapped)
		assert.Len(t, strings.Split(wrapped, "\n"), 2)
	})

	t.Run("big string with the need to line break", func(t *testing.T) {
		wrapped := WordWrap(bigTestString, 100)
		assert.Len(t, strings.Split(wrapped, "\n"), 9)
	})

	t.Run("big string with no need to line break", func(t *testing.T) {
		wrapped := WordWrap(bigTestString, 1000)
		assert.Equal(t, bigTestString, wrapped)
		assert.Len(t, strings.Split(wrapped, "\n"), 1)
	})

	t.Run("simple string with no need to line break", func(t *testing.T) {
		wrapped := WordWrap(simpleTestString, 100)
		assert.Equal(t, simpleTestString, wrapped)
	})

	t.Run("simple string with the need to line break and validate not to break words", func(t *testing.T) {
		wrapped := WordWrap(simpleTestString, 11)
		t.Logf("wrapped: '%#v'", wrapped)
		assert.Len(t, strings.Split(wrapped, "\n"), 6)
		assert.Equal(t, simpleTestStringWrapperd, wrapped)
	})
}
