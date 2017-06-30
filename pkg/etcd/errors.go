package etcd

import (
	"errors"
)

var (
	// ErrKeyAlreadyExists - Testable error for when a key already exists
	ErrKeyAlreadyExists = errors.New("Key Already Exists")

	// ErrKeyMissing - testable error for no expected key defined
	ErrKeyMissing = errors.New("Key not defined")
)
