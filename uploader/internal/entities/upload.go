package entities

import "io"

type File struct {
	UUID      string
	Extension string
	Data      io.ReadCloser
	Size      int64
	Type      string
}
