package cache

import (
	"chat-poc/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cache"
	bolt "go.etcd.io/bbolt"
)

var (
	_ cache.Backend = &BoltDBBackend{}
)

var (
	dbMap = make(map[string]*bolt.DB)
)

type BoltDBBackend struct {
	db *bolt.DB
}

func GetDB(path string) (*bolt.DB, error) {
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
	log := slog.With("pkg", "cache", "key", key)
	log.Debug("BoltDBBackend.Get.Begin")
	bucket, tx, err := b.getBucket(ctx, false)
	if err != nil {
		return nil
	}
	element := bucket.Get([]byte(key))
	if element == nil {
		return nil
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := tx.Commit(); err != nil {
		return nil
	}
	var response llms.ContentResponse
	if err := json.Unmarshal(element, &response); err != nil {
		return nil
	}

	log.With("response", response).Debug("BoltDBBackend.Get.End")

	return &response
}

func (b *BoltDBBackend) Put(ctx context.Context, key string, response *llms.ContentResponse) {
	log := slog.With("pkg", "cache", "key", key, "response", response)
	log.Debug("BoltDBBackend.Put.Begin")
	bucket, tx, err := b.getBucket(ctx, true)
	if err != nil {
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()
	element, _ := json.Marshal(response)
	if err := bucket.Put([]byte(key), element); err != nil {
		log.With("error", err).Error("BoltDBBackend.Put.Error")
		return
	}

	if err := tx.Commit(); err != nil {
		log.With("error", err).Error("BoltDBBackend.Put.Error")
		return
	}
	_ = b.db.Sync()
	log.Debug("BoltDBBackend.Put.End")
}

func (b *BoltDBBackend) List(ctx context.Context) error {
	log := slog.With("pkg", "cache")
	log.Debug("BoltDBBackend.List.Begin")
	bucket, tx, err := b.getBucket(ctx, false)
	if err != nil {
		log.With("error", err).Error("BoltDBBackend.List.Error")
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	cursor := bucket.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		fmt.Printf("- %s\n  %s\n", string(k), string(v))
		//var response llms.ContentResponse
		//if err := json.Unmarshal(v, &response); err != nil {
		//	return err
		//}
	}
	log.Debug("BoltDBBackend.List.End")

	return nil
}

func (b *BoltDBBackend) getBucket(_ context.Context, writable bool) (*bolt.Bucket, *bolt.Tx, error) {
	tx, err := b.db.Begin(writable)
	if err != nil {
		return nil, nil, err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte("response_cache"))
	if err != nil {
		return nil, nil, err
	}
	return bucket, tx, nil
}

func (b *BoltDBBackend) Close() error {
	return b.db.Close()
}
