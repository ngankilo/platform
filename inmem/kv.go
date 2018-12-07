package inmem

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/google/btree"
	"github.com/influxdata/platform/kv"
)

// KVStore is an in memory btree backed kv.Store.
type KVStore struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
}

// NewKVStore creates an instance of a KVStore.
func NewKVStore() *KVStore {
	return &KVStore{
		buckets: map[string]*Bucket{},
	}
}

// View opens up a transaction with a read lock.
func (s *KVStore) View(fn func(kv.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(&Tx{kv: s})
}

// Update opens up a transaction with a write lock.
func (s *KVStore) Update(fn func(kv.Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn(&Tx{kv: s})
}

// Tx is an in memory transaction.
// TODO: make transactions actually transactional
type Tx struct {
	kv *KVStore
}

// CreateBucketIfNotExists creates a btree bucket at the provided key.
func (t *Tx) CreateBucketIfNotExists(b []byte) error {
	// TODO(desa): what to make this value
	t.kv.buckets[string(b)] = &Bucket{btree.New(2)}
	return nil
}

// Bucket retrieves the bucket at the provided key.
func (t *Tx) Bucket(b []byte) (kv.Bucket, error) {
	bkt, ok := t.kv.buckets[string(b)]
	if !ok {
		return nil, errors.New("bucket was not created")
	}

	return bkt, nil
}

// Bucket is a btree that implements kv.Bucket.
type Bucket struct {
	btree *btree.BTree
}

type item struct {
	key   []byte
	value []byte
}

func (i *item) Less(b btree.Item) bool {
	j, ok := b.(*item)
	if !ok {
		return false
	}

	return bytes.Compare(i.key, j.key) < 0
}

// Get retrieves the value at the provided key.
func (b *Bucket) Get(key []byte) ([]byte, error) {
	i := b.btree.Get(&item{key: key})

	if i == nil {
		return nil, kv.ErrKeyNotFound
	}

	j, ok := i.(*item)
	if !ok {
		return nil, fmt.Errorf("error item is type %T not *item", i)
	}

	return j.value, nil
}

// Put sets the key value pair provided.
func (b *Bucket) Put(key []byte, value []byte) error {
	_ = b.btree.ReplaceOrInsert(&item{key: key, value: value})
	return nil
}

// Delete removes the key provided.
func (b *Bucket) Delete(key []byte) error {
	_ = b.btree.Delete(&item{key: key})
	return nil
}

// Cursor creates a static cursor from all entries in the database.
func (b *Bucket) Cursor() (kv.Cursor, error) {
	// TODO we should do this by using the Ascend/Descend methods that
	// the btree provides.
	pairs, err := b.getAll()
	if err != nil {
		return nil, err
	}

	return kv.NewStaticCursor(pairs), nil
}

func (b *Bucket) getAll() ([]kv.Pair, error) {
	pairs := []kv.Pair{}
	var err error
	b.btree.Ascend(func(i btree.Item) bool {
		j, ok := i.(*item)
		if !ok {
			err = fmt.Errorf("error item is type %T not *item", i)
			return false
		}

		pairs = append(pairs, kv.Pair{Key: j.key, Value: j.value})
		return true
	})

	if err != nil {
		return nil, err
	}

	return pairs, nil
}
