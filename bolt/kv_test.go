package bolt_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/platform/bolt"
	"github.com/influxdata/platform/kv"
)

func NewTestKVStore() (*bolt.KVStore, func(), error) {
	f, err := ioutil.TempFile("", "influxdata-platform-bolt-")
	if err != nil {
		return nil, nil, errors.New("unable to open temporary boltdb file")
	}
	f.Close()

	path := f.Name()
	s := bolt.NewKVStore(path)
	if err := s.Open(context.TODO()); err != nil {
		return nil, nil, err
	}

	close := func() {
		s.Close()
		os.Remove(path)
	}

	return s, close, nil
}

// TODO: more tests here for the bolt kv store

func TestKVStore(t *testing.T) {
	s, close, err := NewTestKVStore()
	if err != nil {
		t.Fatalf("failed to create test kv store: %v", err)
	}
	defer close()

	if err := s.Update(func(tx kv.Tx) error {
		if err := tx.CreateBucketIfNotExists([]byte("example")); err != nil {
			t.Fatalf("error creating bucket: %v", err)
		}

		b, err := tx.Bucket([]byte("example"))
		if err != nil {
			t.Fatalf("error getting bucket: %v", err)
		}

		if err := b.Put([]byte("hello"), []byte("world")); err != nil {
			t.Fatalf("error putting kv: %v", err)
		}

		val, err := b.Get([]byte("hello"))
		if err != nil {
			t.Fatalf("error getting value: %v", err)
		}

		if !bytes.Equal(val, []byte("world")) {
			t.Fatalf("error getting value. got %v, expected world", string(val))
		}

		other, err := b.Get([]byte("world"))
		if err != kv.ErrKeyNotFound {
			t.Fatalf("expected key not found error got %v", err)
		}

		if other != nil {
			t.Fatalf("expected other to be nil got %v", string(other))
		}
		return nil
	}); err != nil {
		t.Fatalf("error creating transaction: %v", err)
	}

}
