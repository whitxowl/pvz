package srvErr

import "errors"

var (
	ErrUserExists  = errors.New("user already exists")
	ErrInvalidRole = errors.New("invalid role")
)
