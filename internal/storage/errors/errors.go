package storageErr

import "errors"

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	ErrPVZExists = errors.New("pvz already exists")

	ErrInProgressReceptionExists = errors.New("reception in-progress exists")
	ErrNoInProgressReception     = errors.New("no in-progress reception")
)
