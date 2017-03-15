package etcd

import (
	"errors"
)


var (
	ErrKeyAlreadyExists = errors.New("Key Already Exists")
	ErrKeyMissing = errors.New("Key not defined")
)
