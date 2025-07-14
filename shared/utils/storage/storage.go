package storage

import "errors"

var (
	ErrUUIDExists   = errors.New("UUID already exists")
	ErrUUIDNotFound = errors.New("UUID not found")
	ErrNewerStatus  = errors.New("the file has newer status")
	ErrNoFiles      = errors.New("no files found")
)
