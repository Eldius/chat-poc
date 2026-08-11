package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestPool(t *testing.T) {
	p := NewPool()
	t.Cleanup(func() { _ = p.Close() })

	t.Run("same path returns same handle", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pool.db")

		db1, err := p.Get(path)
		assert.NoError(t, err)

		db2, err := p.Get(path)
		assert.NoError(t, err)
		assert.Same(t, db1, db2)
	})

	t.Run("unopenable path returns error", func(t *testing.T) {
		_, err := p.Get(filepath.Join(t.TempDir(), "missing-dir", "db.db"))
		assert.Error(t, err)
	})

	t.Run("close releases handles", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "close.db")
		_, err := p.Get(path)
		assert.NoError(t, err)
		assert.NoError(t, p.Close())

		// After Close, the path is forgotten and can be reopened.
		db, err := p.Get(path)
		assert.NoError(t, err)
		assert.NotNil(t, db)
	})
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

	entries, err := b.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	assert.Len(t, entries, 2)
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
