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
		{
			name: "Put",
			fn:   KVPut,
		},
		{
			name: "Delete",
			fn:   KVDelete,
		},
		{
			name: "Cursor",
			fn:   KVCursor,
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
					t.Errorf("unexpected error retrieving bucket: %v", err)
					return err
				}

				val, err := b.Get(tt.args.key)
				if (err != nil) != (tt.wants.err != nil) {
					t.Errorf("expected error '%v' got '%v'", tt.wants.err, err)
					return err
				}

				if err != nil && tt.wants.err != nil {
					if err.Error() != tt.wants.err.Error() {
						t.Errorf("expected error messages to match '%v' got '%v'", tt.wants.err, err.Error())
						return err
					}
				}

				if want, got := tt.wants.val, val; !bytes.Equal(want, got) {
					t.Errorf("exptected to get value %s got %s", string(want), string(got))
					return err
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
					t.Errorf("unexpected error retrieving bucket: %v", err)
					return err
				}

				{
					err := b.Put(tt.args.key, tt.args.val)
					if (err != nil) != (tt.wants.err != nil) {
						t.Errorf("expected error '%v' got '%v'", tt.wants.err, err)
						return err
					}

					if err != nil && tt.wants.err != nil {
						if err.Error() != tt.wants.err.Error() {
							t.Errorf("expected error messages to match '%v' got '%v'", tt.wants.err, err.Error())
							return err
						}
					}

					val, err := b.Get(tt.args.key)
					if err != nil {
						t.Errorf("unexpected error retrieving value: %v", err)
						return err
					}

					if want, got := tt.args.val, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
						return err
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

// KVDelete tests the delete method contract for the key value store.
func KVDelete(
	init func(KVStoreFields, *testing.T) (kv.Store, func()),
	t *testing.T,
) {
	type args struct {
		bucket []byte
		key    []byte
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
			name: "delete key",
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
					t.Errorf("unexpected error retrieving bucket: %v", err)
					return err
				}

				{
					err := b.Delete(tt.args.key)
					if (err != nil) != (tt.wants.err != nil) {
						t.Errorf("expected error '%v' got '%v'", tt.wants.err, err)
						return err
					}

					if err != nil && tt.wants.err != nil {
						if err.Error() != tt.wants.err.Error() {
							t.Errorf("expected error messages to match '%v' got '%v'", tt.wants.err, err.Error())
							return err
						}
					}

					if _, err := b.Get(tt.args.key); err != kv.ErrKeyNotFound {
						t.Errorf("expected key not found error got %v", err)
						return err
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

// KVCursor tests the cursor contract for the key value store.
func KVCursor(
	init func(KVStoreFields, *testing.T) (kv.Store, func()),
	t *testing.T,
) {
	type args struct {
		bucket []byte
		seek   []byte
	}
	type wants struct {
		err   error
		first kv.Pair
		last  kv.Pair
		seek  kv.Pair
		next  kv.Pair
		prev  kv.Pair
	}

	tests := []struct {
		name   string
		fields KVStoreFields
		args   args
		wants  wants
	}{
		{
			name: "basic cursor",
			fields: KVStoreFields{
				Bucket: []byte("bucket"),
				Pairs: []kv.Pair{
					{
						Key:   []byte("a"),
						Value: []byte("1"),
					},
					{
						Key:   []byte("ab"),
						Value: []byte("2"),
					},
					{
						Key:   []byte("abc"),
						Value: []byte("3"),
					},
					{
						Key:   []byte("abcd"),
						Value: []byte("4"),
					},
					{
						Key:   []byte("abcde"),
						Value: []byte("5"),
					},
					{
						Key:   []byte("bcd"),
						Value: []byte("6"),
					},
					{
						Key:   []byte("cd"),
						Value: []byte("7"),
					},
				},
			},
			args: args{
				bucket: []byte("bucket"),
				seek:   []byte("abc"),
			},
			wants: wants{
				first: kv.Pair{
					Key:   []byte("a"),
					Value: []byte("1"),
				},
				last: kv.Pair{
					Key:   []byte("cd"),
					Value: []byte("7"),
				},
				seek: kv.Pair{
					Key:   []byte("abc"),
					Value: []byte("3"),
				},
				next: kv.Pair{
					Key:   []byte("abcd"),
					Value: []byte("4"),
				},
				prev: kv.Pair{
					Key:   []byte("abc"),
					Value: []byte("3"),
				},
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
					t.Errorf("unexpected error retrieving bucket: %v", err)
					return err
				}

				cur, err := b.Cursor()
				if (err != nil) != (tt.wants.err != nil) {
					t.Errorf("expected error '%v' got '%v'", tt.wants.err, err)
					return err
				}

				if err != nil && tt.wants.err != nil {
					if err.Error() != tt.wants.err.Error() {
						t.Errorf("expected error messages to match '%v' got '%v'", tt.wants.err, err.Error())
						return err
					}
				}

				{
					key, val, err := cur.First()
					if err != nil {
						t.Errorf("unexpected error in call to first: %v", err)
						return err
					}
					if want, got := tt.wants.first.Key, key; !bytes.Equal(want, got) {
						t.Errorf("exptected to get key %s got %s", string(want), string(got))
						return err
					}

					if want, got := tt.wants.first.Value, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
						return err
					}
				}

				{
					key, val, err := cur.Last()
					if err != nil {
						t.Errorf("unexpected error in call to last: %v", err)
						return err
					}
					if want, got := tt.wants.last.Key, key; !bytes.Equal(want, got) {
						t.Errorf("exptected to get key %s got %s", string(want), string(got))
						return err
					}

					if want, got := tt.wants.last.Value, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
						return err
					}
				}

				{
					key, val, err := cur.Seek(tt.args.seek)
					if err != nil {
						t.Errorf("unexpected error in call to seek: %v", err)
						return err
					}
					if want, got := tt.wants.seek.Key, key; !bytes.Equal(want, got) {
						t.Errorf("exptected to get key %s got %s", string(want), string(got))
						return err
					}

					if want, got := tt.wants.seek.Value, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
						return err
					}
				}

				{
					key, val, err := cur.Next()
					if err != nil {
						t.Errorf("unexpected error in call to next: %v", err)
						return err
					}
					if want, got := tt.wants.next.Key, key; !bytes.Equal(want, got) {
						t.Errorf("exptected to get key %s got %s", string(want), string(got))
						return err
					}

					if want, got := tt.wants.next.Value, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
						return err
					}
				}

				{
					key, val, err := cur.Prev()
					if err != nil {
						t.Errorf("unexpected error in call to prev: %v", err)
						return err
					}
					if want, got := tt.wants.prev.Key, key; !bytes.Equal(want, got) {
						t.Errorf("exptected to get key %s got %s", string(want), string(got))
						return err
					}

					if want, got := tt.wants.prev.Value, val; !bytes.Equal(want, got) {
						t.Errorf("exptected to get value %s got %s", string(want), string(got))
						return err
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
