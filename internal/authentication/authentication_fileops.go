package authentication

import "os"

type FileOperations struct{}

//go:generate mockery --name=FileInterface --output=../../mocks --outpkg=mocks --filename=fileops.go
type FileInterface interface {
	Open(filename string) (*os.File, error)
	Read(filename string) ([]byte, error)
	Create(filename string) (*os.File, error)
	Write(filename string, data []byte, perm os.FileMode) error
}

func (f *FileOperations) Open(filename string) (*os.File, error) {
	return os.Open(filename)
}

func (f *FileOperations) Read(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (f *FileOperations) Create(filename string) (*os.File, error) {
	return os.Create(filename)
}

func (f *FileOperations) Write(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}
