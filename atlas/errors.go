package atlas

import "errors"

var (
	ErrInvalidTag       = errors.New("tag is not valid")
	ErrResourceNotFound = errors.New("resource does not exist")
	ErrUploadToRoot     = errors.New("cannot write to root")
	ErrNotFile          = errors.New("expecting file, found folder")
	ErrNotFolder        = errors.New("expecting folder, found file")
)
