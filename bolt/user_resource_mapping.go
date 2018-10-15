package bolt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/bbolt"
	"github.com/influxdata/platform"
)

var (
	userResourceMappingBucket = []byte("userresourcemappingsv1")
)

func (c *Client) initializeUserResourceMappings(ctx context.Context, tx *bolt.Tx) error {
	if _, err := tx.CreateBucketIfNotExists([]byte(userResourceMappingBucket)); err != nil {
		return err
	}
	return nil
}

func filterMappingsFn(filter platform.UserResourceMappingFilter) func(m *platform.UserResourceMapping) bool {
	return func(mapping *platform.UserResourceMapping) bool {
		return (!filter.UserID.Valid() || (filter.UserID == mapping.UserID)) &&
			(!filter.ResourceID.Valid() || (filter.ResourceID == mapping.ResourceID)) &&
			(filter.UserType == "" || (filter.UserType == mapping.UserType)) &&
			(filter.ResourceType == "" || (filter.ResourceType == mapping.ResourceType))
	}
}

func (c *Client) FindUserResourceMappings(ctx context.Context, filter platform.UserResourceMappingFilter, opt ...platform.FindOptions) ([]*platform.UserResourceMapping, int, error) {
	ms := []*platform.UserResourceMapping{}
	err := c.db.View(func(tx *bolt.Tx) error {
		mappings, err := c.findUserResourceMappings(ctx, tx, filter)
		if err != nil {
			return err
		}
		ms = mappings
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return ms, len(ms), nil
}

func (c *Client) findUserResourceMappings(ctx context.Context, tx *bolt.Tx, filter platform.UserResourceMappingFilter) ([]*platform.UserResourceMapping, error) {
	ms := []*platform.UserResourceMapping{}
	filterFn := filterMappingsFn(filter)
	err := c.forEachUserResourceMapping(ctx, tx, func(m *platform.UserResourceMapping) bool {
		if filterFn(m) {
			ms = append(ms, m)
		}
		return true
	})

	if err != nil {
		return nil, err
	}

	return ms, nil
}

func (c *Client) findUserResourceMapping(ctx context.Context, tx *bolt.Tx, resourceID platform.ID, userID platform.ID) (*platform.UserResourceMapping, error) {
	var m platform.UserResourceMapping

	key, err := userResourceKey(&platform.UserResourceMapping{
		ResourceID: resourceID,
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	v := tx.Bucket(userResourceMappingBucket).Get(key)
	if len(v) == 0 {
		return nil, fmt.Errorf("userResource mapping not found")
	}

	if err := json.Unmarshal(v, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (c *Client) FindDeepUserResourceMappings(ctx context.Context, filter platform.DeepUserResourceMappingFilter) ([]*platform.UserResourceMapping, int, error) {
	ms := []*platform.UserResourceMapping{}
	err := c.db.View(func(tx *bolt.Tx) error {
		// get resources that belong to user
		userMappingFilter := platform.UserResourceMappingFilter{
			ResourceType: filter.ResourceType,
			UserID:       filter.UserID,
		}

		userMappings, err := c.findUserResourceMappings(ctx, tx, userMappingFilter)
		if err != nil {
			return err
		}

		// TODO (jm): memoize mapping results
		ms = userMappings

		// get resources that belong to user's orgs
		orgMappingFilter := platform.UserResourceMappingFilter{
			ResourceType: platform.OrgResourceType,
			UserID:       filter.UserID,
		}

		orgMappings, err := c.findUserResourceMappings(ctx, tx, orgMappingFilter)
		if err != nil {
			return err
		}

		for _, m := range orgMappings {
			f := platform.UserResourceMappingFilter{
				ResourceType: filter.ResourceType,
				UserID:       m.ResourceID, // the user's organization
			}

			mappings, err := c.findUserResourceMappings(ctx, tx, f)
			if err != nil {
				return err
			}
			ms = append(ms, mappings...)
		}

		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	return ms, len(ms), nil
}

func (c *Client) CreateUserResourceMapping(ctx context.Context, m *platform.UserResourceMapping) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		unique := c.uniqueUserResourceMapping(ctx, tx, m)

		if !unique {
			return fmt.Errorf("mapping %s:%s already exists", m.UserID.String(), m.ResourceID.String())
		}

		v, err := json.Marshal(m)
		if err != nil {
			return err
		}

		key, err := userResourceKey(m)
		if err != nil {
			return err
		}

		if err := tx.Bucket(userResourceMappingBucket).Put(key, v); err != nil {
			return err
		}

		return nil
	})
}

func userResourceKey(m *platform.UserResourceMapping) ([]byte, error) {
	encodedResourceID, err := m.ResourceID.Encode()
	if err != nil {
		return nil, err
	}

	encodedUserID, err := m.UserID.Encode()
	if err != nil {
		return nil, err
	}

	key := make([]byte, len(encodedResourceID)+len(encodedUserID))
	copy(key, encodedResourceID)
	copy(key[len(encodedResourceID):], encodedUserID)

	return key, nil
}

func (c *Client) forEachUserResourceMapping(ctx context.Context, tx *bolt.Tx, fn func(*platform.UserResourceMapping) bool) error {
	cur := tx.Bucket(userResourceMappingBucket).Cursor()
	for k, v := cur.First(); k != nil; k, v = cur.Next() {
		m := &platform.UserResourceMapping{}
		if err := json.Unmarshal(v, m); err != nil {
			return err
		}
		if !fn(m) {
			break
		}
	}

	return nil
}

func (c *Client) uniqueUserResourceMapping(ctx context.Context, tx *bolt.Tx, m *platform.UserResourceMapping) bool {
	key, err := userResourceKey(m)
	if err != nil {
		return false
	}

	v := tx.Bucket(userResourceMappingBucket).Get(key)
	return len(v) == 0
}

func (c *Client) DeleteUserResourceMapping(ctx context.Context, resourceID platform.ID, userID platform.ID) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return c.deleteUserResourceMapping(ctx, tx, resourceID, userID)
	})
}

func (c *Client) deleteUserResourceMapping(ctx context.Context, tx *bolt.Tx, resourceID platform.ID, userID platform.ID) error {
	m, err := c.findUserResourceMapping(ctx, tx, resourceID, userID)
	if err != nil {
		return err
	}

	key, err := userResourceKey(m)
	if err != nil {
		return err
	}

	return tx.Bucket(userResourceMappingBucket).Delete(key)
}
