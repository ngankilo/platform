package bolt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/influxdata/platform/kv"
	"go.uber.org/zap"
)

type KVStore struct {
	path   string
	db     *bolt.DB
	logger *zap.Logger
}

func NewKVStore(path string) *KVStore {
	return &KVStore{
		path:   path,
		logger: zap.NewNop(),
	}
}

// Open / create boltDB file.
func (s *KVStore) Open(ctx context.Context) error {
	// Ensure the required directory structure exists.
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return fmt.Errorf("unable to create directory %s: %v", s.path, err)
	}

	if _, err := os.Stat(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Open database file.
	db, err := bolt.Open(s.path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("unable to open boltdb file %v", err)
	}
	s.db = db

	s.logger.Info("Resources opened", zap.String("path", s.path))
	return nil
}

// Close the connection to the bolt database
func (s *KVStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *KVStore) WithLogger(l *zap.Logger) {
	s.logger = l
}

func (s *KVStore) WithDB(db *bolt.DB) {
	s.db = db
}

func (s *KVStore) View(fn func(tx kv.Tx) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

func (s *KVStore) Update(fn func(tx kv.Tx) error) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{tx})
	})
}

func (s *KVStore) Tx(writable bool) (kv.Tx, error) {
	tx, err := s.db.Begin(writable)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx: tx,
	}, nil

}

type Tx struct {
	tx *bolt.Tx
}

func (tx *Tx) CreateBucketIfNotExists(b []byte) error {
	_, err := tx.tx.CreateBucketIfNotExists(b)
	if err != nil {
		return err
	}
	return nil
}

func (tx *Tx) Bucket(b []byte) (kv.Bucket, error) {
	return &Bucket{
		bucket: tx.tx.Bucket(b),
	}, nil
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}

type Bucket struct {
	bucket *bolt.Bucket
}

func (b *Bucket) Get(key []byte) ([]byte, error) {
	val := b.bucket.Get(key)
	if len(val) == 0 {
		return nil, kv.ErrKeyNotFound
	}

	return val, nil
}

func (b *Bucket) Put(key []byte, value []byte) error {
	return b.bucket.Put(key, value)
}

func (b *Bucket) Delete(key []byte) error {
	return b.bucket.Delete(key)
}

func (b *Bucket) Cursor() (kv.Cursor, error) {
	return &Cursor{
		cursor: b.bucket.Cursor(),
	}, nil
}

type Cursor struct {
	cursor *bolt.Cursor
}

func (c *Cursor) Seek(prefix []byte) ([]byte, []byte, error) {
	k, v := c.cursor.Seek(prefix)
	// TODO(desa): what possible errors?
	return k, v, nil
}

func (c *Cursor) First() ([]byte, []byte, error) {
	k, v := c.cursor.First()
	// TODO(desa): what possible errors?
	return k, v, nil
}

func (c *Cursor) Last() ([]byte, []byte, error) {
	k, v := c.cursor.Last()
	// TODO(desa): what possible errors?
	return k, v, nil
}

func (c *Cursor) Next() ([]byte, []byte, error) {
	k, v := c.cursor.Next()
	// TODO(desa): what possible errors?
	return k, v, nil
}

func (c *Cursor) Prev() ([]byte, []byte, error) {
	k, v := c.cursor.Prev()
	// TODO(desa): what possible errors?
	return k, v, nil
}
