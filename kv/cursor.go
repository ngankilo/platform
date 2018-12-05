package kv

import (
	"bytes"
	"fmt"
	"sort"
)

// StaticCursor implements the Cursor interface for a slice of
// static key value pairs.
type StaticCursor struct {
	idx   int
	pairs []Pair
}

// Pair is a struct for key value pairs.
type Pair struct {
	Key   []byte
	Value []byte
}

// NewStaticCursor returns an instance of a StaticCursor. It
// destructively sorts the provided pairs to be in key ascending order.
func NewStaticCursor(pairs []Pair) *StaticCursor {
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].Key, pairs[j].Key) < 0
	})
	return &StaticCursor{
		pairs: pairs,
	}
}

// Seek searches the slice for the first key with the provided prefix.
func (c *StaticCursor) Seek(prefix []byte) ([]byte, []byte, error) {
	// TODO: do binary search for prefix since pairs are ordered.
	for i, pair := range c.pairs {
		if bytes.HasPrefix(pair.Key, prefix) {
			c.idx = i
			return pair.Key, pair.Value, nil
		}
	}

	return nil, nil, fmt.Errorf("prefix not found")
}

func (c *StaticCursor) getValueAtIndex() ([]byte, []byte, error) {
	if c.idx < 0 {
		return nil, nil, fmt.Errorf("index value is negative")
	}

	if c.idx >= len(c.pairs) {
		return nil, nil, fmt.Errorf("index exceeds the length of the pairs")
	}

	pair := c.pairs[c.idx]

	return pair.Key, pair.Value, nil
}

// First retrieves the first element in the cursor.
func (c *StaticCursor) First() ([]byte, []byte, error) {
	c.idx = 0
	return c.getValueAtIndex()
}

// Last retrieves the last element in the cursor.
func (c *StaticCursor) Last() ([]byte, []byte, error) {
	c.idx = len(c.pairs) - 1
	return c.getValueAtIndex()
}

// Next retrieves the next entry in the cursor.
func (c *StaticCursor) Next() ([]byte, []byte, error) {
	c.idx += 1
	return c.getValueAtIndex()
}

// Prev retrieves the previous entry in the cursor.
func (c *StaticCursor) Prev() ([]byte, []byte, error) {
	c.idx -= 1
	return c.getValueAtIndex()
}
