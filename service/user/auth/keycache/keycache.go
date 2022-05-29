package keycache

import (
	"errors"
)

var keyCache map[string][]byte = map[string][]byte{}

func GetUserKey(uuid string) ([]byte, error) {
	if val, ok := keyCache[uuid]; ok {
		return val, nil
	}

	return nil, errors.New("User has no cached key")
}

func SetUserKey(uuid string, key []byte) error {
	keyCache[uuid] = key

	return nil
}

func Cache() map[string][]byte {
	return keyCache
}
