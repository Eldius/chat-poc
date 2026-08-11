package cache

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms"
	bolt "go.etcd.io/bbolt"
)

func tempDB(t *testing.T) *bolt.DB {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	db, err := bolt.Open(f.Name(), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestBoltDBBackend_PutAndGet(t *testing.T) {
	db := tempDB(t)
	b := NewBoltDBBackend(db)

	ctx := context.Background()
	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: "hello world"},
		},
	}

	b.Put(ctx, "key1", resp)
	got := b.Get(ctx, "key1")
	assert.NotNil(t, got, "expected non-nil response")
	assert.Greater(t, len(got.Choices), 0, "expected at least one choice")
	assert.Equal(t, "hello world", got.Choices[0].Content, "expected 'hello world', got '%s'", got.Choices[0].Content)
}

func TestBoltDBBackend_GetMissingKey(t *testing.T) {
	db := tempDB(t)
	b := NewBoltDBBackend(db)

	got := b.Get(context.Background(), "nonexistent")
	if got != nil {
		t.Fatal("expected nil for missing key")
	}
}

func TestBoltDBBackend_List(t *testing.T) {
	db := tempDB(t)
	b := NewBoltDBBackend(db)

	ctx := context.Background()
	b.Put(ctx, "a", &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "1"}}})
	b.Put(ctx, "b", &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "2"}}})

	if err := b.List(ctx); err != nil {
		t.Fatalf("List failed: %v", err)
	}
}

func TestBoltDBBackend_OverrideValue(t *testing.T) {
	db := tempDB(t)
	b := NewBoltDBBackend(db)

	ctx := context.Background()
	b.Put(ctx, "k", &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "first"}}})
	b.Put(ctx, "k", &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "second"}}})

	got := b.Get(ctx, "k")
	if got == nil || len(got.Choices) == 0 {
		t.Fatal("expected non-nil response")
	}
	if got.Choices[0].Content != "second" {
		t.Fatalf("expected 'second', got '%s'", got.Choices[0].Content)
	}
}
