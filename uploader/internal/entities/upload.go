package entities

import "io"

type File struct {
	_uuid      string
	_extension string
	_data      io.ReadCloser
	_size      int64
	_type      string
}

func New(u, ext string, d io.ReadCloser, s int64, t string) File {
	return File{
		_uuid:      u,
		_extension: ext,
		_data:      d,
		_size:      s,
		_type:      t,
	}
}

func (f File) FullName() string {
	return f._uuid + f._extension
}

func (f File) Data() io.Reader {
	return f._data
}

func (f File) Size() int64 {
	return f._size
}

func (f File) MimeType() string {
	return f._type
}

func (f File) Close() error {
	return f._data.Close()
}
