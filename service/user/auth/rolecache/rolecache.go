package rolecache

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/tidwall/buntdb"
)

// RoleCache is a cache for user roles
type RoleCache struct {
	GetUserRoles func(userID string) ([]string, error)
	DB           *buntdb.DB
}

func makeKey(userID string) string {
	return userID
}

// Invalidate invalidates an user's roles
func (r *RoleCache) Invalidate(userID string) error {
	key := makeKey(userID)
	err := r.DB.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		if err == buntdb.ErrNotFound {
			return nil
		}
		return err
	})

	return err
}

// GetRoles returns an user's roles
func (r *RoleCache) GetRoles(userID string) ([]string, error) {
	roles := []string{}
	key := makeKey(userID)
	err := r.DB.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(val), &roles)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && err == buntdb.ErrNotFound {
		newRoles, err := r.GetUserRoles(userID)
		if err != nil {
			return roles, errors.Wrap(err, "Couldn't GetUserRoles for "+userID)
		}

		_ = r.updateRoles(userID, newRoles)

		return newRoles, err
	}

	return roles, err
}

// updateRoles updates an user's roles in the cache
func (r *RoleCache) updateRoles(userID string, roles []string) error {
	key := makeKey(userID)
	byteVal, err := json.Marshal(&roles)
	if err != nil {
		return errors.Wrap(err, "Error marshalling roles")
	}

	err = r.DB.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, string(byteVal), nil)
		return err
	})

	return err
}
