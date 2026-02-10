package storageErr

import "errors"

var (
	ErrUserExists = errors.New("user already exists")
)
