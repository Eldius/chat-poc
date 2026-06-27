package cache

import (
	"chat-poc/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cache"
	bolt "go.etcd.io/bbolt"
)

var (
	_ cache.Backend = &BoltDBBackend{}
)

var (
	dbMap   = make(map[string]*bolt.DB)
	dbMapMu sync.RWMutex
)

type BoltDBBackend struct {
	db *bolt.DB
}

func GetDB(path string) (*bolt.DB, error) {
	dbMapMu.RLock()
	db, ok := dbMap[path]
	dbMapMu.RUnlock()
	if ok {
		return db, nil
	}

	dbMapMu.Lock()
	defer dbMapMu.Unlock()

	if db, ok := dbMap[path]; ok {
		return db, nil
	}

	opts := bolt.DefaultOptions
	opts.ReadOnly = false
	opts.Timeout = config.GetBedrockCachePersistTimeout()

	db, err := bolt.Open(path, 0600, opts)
	if err != nil {
		return nil, err
	}
	dbMap[path] = db
	return db, nil
}

func NewBoltDBBackend(db *bolt.DB) *BoltDBBackend {
	return &BoltDBBackend{db: db}
}

func (b *BoltDBBackend) Get(ctx context.Context, key string) *llms.ContentResponse {
	tx, err := b.db.Begin(false)
	if err != nil {
		return nil
	}
	defer func() {
		_ = tx.Rollback()
	}()

	bucket := tx.Bucket([]byte("response_cache"))
	if bucket == nil {
		return nil
	}

	element := bucket.Get([]byte(key))
	if element == nil {
		return nil
	}

	var response llms.ContentResponse
	if err := json.Unmarshal(element, &response); err != nil {
		return nil
	}

	return &response
}

func (b *BoltDBBackend) Put(ctx context.Context, key string, response *llms.ContentResponse) {
	bucket, tx, err := b.getBucket(ctx, true)
	if err != nil {
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	element, _ := json.Marshal(response)
	if err := bucket.Put([]byte(key), element); err != nil {
		return
	}

	if err := tx.Commit(); err != nil {
		return
	}
	if err := b.db.Sync(); err != nil {
		return
	}
}

func (b *BoltDBBackend) List(ctx context.Context) error {
	tx, err := b.db.Begin(false)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	bucket := tx.Bucket([]byte("response_cache"))
	if bucket == nil {
		return nil
	}

	cursor := bucket.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		fmt.Printf("- %s\n  %s\n", string(k), string(v))
	}

	return nil
}

func (b *BoltDBBackend) getBucket(_ context.Context, writable bool) (*bolt.Bucket, *bolt.Tx, error) {
	tx, err := b.db.Begin(writable)
	if err != nil {
		return nil, nil, err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte("response_cache"))
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	return bucket, tx, nil
}

func (b *BoltDBBackend) Close() error {
	return b.db.Close()
}
