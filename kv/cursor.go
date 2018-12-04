package kv

import (
	"bytes"
	"fmt"
)

type StaticCursor struct {
	idx   int
	pairs []Pair
}

func NewStaticCursor(pairs []Pair) *StaticCursor {
	return &StaticCursor{
		pairs: pairs,
	}
}

func (c *StaticCursor) Seek(prefix []byte) ([]byte, []byte, error) {
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

func (c *StaticCursor) First() ([]byte, []byte, error) {
	c.idx = 0
	return c.getValueAtIndex()
}

func (c *StaticCursor) Last() ([]byte, []byte, error) {
	c.idx = len(c.pairs) - 1
	return c.getValueAtIndex()
}

func (c *StaticCursor) Next() ([]byte, []byte, error) {
	c.idx += 1
	return c.getValueAtIndex()
}

func (c *StaticCursor) Prev() ([]byte, []byte, error) {
	c.idx -= 1
	return c.getValueAtIndex()
}
