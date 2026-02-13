package srvErr

import "errors"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrPVZExists   = errors.New("pvz already exists")
	ErrInvalidCity = errors.New("invalid city")

	ErrInProgressReceptionExists = errors.New("in-progress reception exists")
	ErrNoInProgressReception     = errors.New("no in-progress reception exists")
)
