package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cache"
	bolt "go.etcd.io/bbolt"
)

var (
	_ cache.Backend = &BoltDBBackend{}
)

const bucketName = "response_cache"

// Pool manages shared open BoltDB handles keyed by file path.
type Pool struct {
	mu  sync.RWMutex
	dbs map[string]*bolt.DB
}

func NewPool() *Pool {
	return &Pool{dbs: make(map[string]*bolt.DB)}
}

// Get returns the pooled handle for path, opening it on first use.
func (p *Pool) Get(path string) (*bolt.DB, error) {
	p.mu.RLock()
	db, ok := p.dbs[path]
	p.mu.RUnlock()
	if ok {
		return db, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if db, ok := p.dbs[path]; ok {
		return db, nil
	}

	db, err := bolt.Open(path, 0600, bolt.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("opening bolt db %q: %w", path, err)
	}
	p.dbs[path] = db
	return db, nil
}

// Close closes all pooled handles.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for path, db := range p.dbs {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing bolt db %q: %w", path, err))
		}
		delete(p.dbs, path)
	}
	return errors.Join(errs...)
}

type BoltDBBackend struct {
	db *bolt.DB
}

func NewBoltDBBackend(db *bolt.DB) *BoltDBBackend {
	return &BoltDBBackend{db: db}
}

func (b *BoltDBBackend) Get(ctx context.Context, key string) *llms.ContentResponse {
	tx, err := b.db.Begin(false)
	if err != nil {
		slog.WarnContext(ctx, "cache get: beginning tx", "error", err)
		return nil
	}
	defer func() {
		_ = tx.Rollback()
	}()

	bucket := tx.Bucket([]byte(bucketName))
	if bucket == nil {
		return nil
	}

	element := bucket.Get([]byte(key))
	if element == nil {
		return nil
	}

	var response llms.ContentResponse
	if err := json.Unmarshal(element, &response); err != nil {
		slog.WarnContext(ctx, "cache get: unmarshaling entry", "key", key, "error", err)
		return nil
	}

	return &response
}

// Put stores a response. The langchaingo cache.Backend interface has no error
// return, so failures are logged and treated as a cache skip.
func (b *BoltDBBackend) Put(ctx context.Context, key string, response *llms.ContentResponse) {
	bucket, tx, err := b.getBucket(true)
	if err != nil {
		slog.WarnContext(ctx, "cache put: opening bucket", "error", err)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	element, err := json.Marshal(response)
	if err != nil {
		slog.WarnContext(ctx, "cache put: marshaling response", "key", key, "error", err)
		return
	}
	if err := bucket.Put([]byte(key), element); err != nil {
		slog.WarnContext(ctx, "cache put: writing entry", "key", key, "error", err)
		return
	}
	if err := tx.Commit(); err != nil {
		slog.WarnContext(ctx, "cache put: committing tx", "key", key, "error", err)
		return
	}
	if err := b.db.Sync(); err != nil {
		slog.WarnContext(ctx, "cache put: syncing db", "key", key, "error", err)
	}
}

// List returns all cached entries.
func (b *BoltDBBackend) List(ctx context.Context) ([]Entry, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	bucket := tx.Bucket([]byte(bucketName))
	if bucket == nil {
		return nil, nil
	}

	var entries []Entry
	cursor := bucket.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		var response llms.ContentResponse
		if err := json.Unmarshal(v, &response); err != nil {
			return nil, fmt.Errorf("unmarshaling cache entry %q: %w", string(k), err)
		}
		entries = append(entries, Entry{Key: string(k), Response: response})
	}

	return entries, nil
}

// Entry is a single cached request/response pair.
type Entry struct {
	Key      string
	Response llms.ContentResponse
}

func (b *BoltDBBackend) getBucket(writable bool) (*bolt.Bucket, *bolt.Tx, error) {
	tx, err := b.db.Begin(writable)
	if err != nil {
		return nil, nil, err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	return bucket, tx, nil
}

func (b *BoltDBBackend) Close() error {
	return b.db.Close()
}
