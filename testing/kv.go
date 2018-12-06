package testing

import (
	"bytes"
	"testing"

	"github.com/influxdata/platform/kv"
)

// KVStoreFields are background data that has to be set before
// the test runs.
type KVStoreFields struct {
	Bucket []byte
	Pairs  []kv.Pair
}

// KVStore tests the key value store contract
func KVStore(
	init func(KVStoreFields, *testing.T) (kv.Store, func()),
	t *testing.T,
) {
	tests := []struct {
		name string
		fn   func(
			init func(KVStoreFields, *testing.T) (kv.Store, func()),
			t *testing.T,
		)
	}{
		{
			name: "Get",
			fn:   KVGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(init, t)
		})
	}
}

// KVGet tests the get method contract for the key value store.
func KVGet(
	init func(KVStoreFields, *testing.T) (kv.Store, func()),
	t *testing.T,
) {
	type args struct {
		bucket []byte
		key    []byte
	}
	type wants struct {
		err error
		val []byte
	}

	tests := []struct {
		name   string
		fields KVStoreFields
		args   args
		wants  wants
	}{
		{
			name: "get key",
			fields: KVStoreFields{
				Bucket: []byte("bucket"),
				Pairs: []kv.Pair{
					{
						Key:   []byte("hello"),
						Value: []byte("world"),
					},
				},
			},
			args: args{
				bucket: []byte("bucket"),
				key:    []byte("hello"),
			},
			wants: wants{
				val: []byte("world"),
			},
		},
		{
			name: "get missing key",
			fields: KVStoreFields{
				Bucket: []byte("bucket"),
				Pairs:  []kv.Pair{},
			},
			args: args{
				bucket: []byte("bucket"),
				key:    []byte("hello"),
			},
			wants: wants{
				err: kv.ErrKeyNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, close := init(tt.fields, t)
			defer close()

			err := s.View(func(tx kv.Tx) error {
				b, err := tx.Bucket(tt.args.bucket)
				if err != nil {
					t.Fatalf("unexpected error retrieving bucket: %v", err)
				}

				val, err := b.Get(tt.args.key)
				if (err != nil) != (tt.wants.err != nil) {
					t.Fatalf("expected error '%v' got '%v'", tt.wants.err, err)
				}

				if err != nil && tt.wants.err != nil {
					if err.Error() != tt.wants.err.Error() {
						t.Fatalf("expected error messages to match '%v' got '%v'", tt.wants.err, err.Error())
					}
				}

				if want, got := tt.wants.val, val; !bytes.Equal(want, got) {
					t.Errorf("exptected to get value %s got %s", string(want), string(got))
				}

				return nil
			})

			if err != nil {
				t.Fatalf("error during view transaction: %v", err)
			}
		})
	}
}

// KVPut tests the get method contract for the key value store.
func KVPut(
	init func(KVStoreFields, *testing.T) (kv.Store, func()),
	t *testing.T,
) {
	type args struct {
		bucket []byte
		key    []byte
		val    []byte
	}
	type wants struct {
		err error
	}

	tests := []struct {
		name   string
		fields KVStoreFields
		args   args
		wants  wants
	}{
		{
			name: "put pair",
			fields: KVStoreFields{
				Bucket: []byte("bucket"),
				Pairs:  []kv.Pair{},
			},
			args: args{
				bucket: []byte("bucket"),
				key:    []byte("hello"),
				val:    []byte("world"),
			},
			wants: wants{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, close := init(tt.fields, t)
			defer close()

			err := s.Update(func(tx kv.Tx) error {
				b, err := tx.Bucket(tt.args.bucket)
				if err != nil {
					t.Fatalf("unexpected error retrieving bucket: %v", err)
				}

				{
					err := b.Put(tt.args.key, tt.args.val)
					if (err != nil) != (tt.wants.err != nil) {
						t.Fatalf("expected error '%v' got '%v'", tt.wants.err, err)
					}

					if err != nil && tt.wants.err != nil {
						if err.Error() != tt.wants.err.Error() {
							t.Fatalf("expected error messages to match '%v' got '%v'", tt.wants.err, err.Error())
						}
					}

					val, err := b.Get(tt.args.key)
					if err != nil {
						t.Fatalf("unexpected error retrieving value: %v", err)
					}

					if want, got := tt.args.val, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
					}
				}

				return nil
			})

			if err != nil {
				t.Fatalf("error during view transaction: %v", err)
			}
		})
	}
}
