package kv

import (
	"fmt"
)

var (
	ErrKeyNotFound = fmt.Errorf("key not found")
)

type Store interface {
	Tx(bool) (Tx, error)
	View(func(Tx) error) error
	Update(func(Tx) error) error
}

type Tx interface {
	CreateBucketIfNotExists(b []byte) error
	Bucket(b []byte) (Bucket, error)
	Commit() error
	Rollback() error
}

type Bucket interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	Cursor() (Cursor, error)
}

type Cursor interface {
	Seek(prefix []byte) (k []byte, v []byte, err error)
	First() (k []byte, v []byte, err error)
	Last() (k []byte, v []byte, err error)
	Next() (k []byte, v []byte, err error)
	Prev() (k []byte, v []byte, err error)
}
