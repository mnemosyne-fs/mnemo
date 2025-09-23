package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const filePerm = 0666

var tag_validate = regexp.MustCompile(`^[\w\-. ]+$`)
var ErrInvalidTag = errors.New("tag is not valid")

func ValidateTag(tag string) error {
	if !tag_validate.Match([]byte(tag)) {
		return ErrInvalidTag
	}
	return nil
}

func ValidatePath(path string) error {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) {
		return fmt.Errorf("absolute paths are not allowed: %s", path)
	}

	rel, err := filepath.Rel(".", clean)
	if err != nil {
		return fmt.Errorf("cannot evaluate path: %w", err)
	}

	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path escapes base directory: %s", path)
	}

	return nil
}

type Path string

func NewPath(path string) Path {
	return Path(path)
}

func (p *Path) Resolve(atlas *Atlas) string {
	return filepath.Join(atlas.root, string(*p))
}

func (p *Path) Stat(atlas *Atlas) (os.FileInfo, error) {
	info, err := os.Stat(p.Resolve(atlas))
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (p *Path) Validate() error {
	return ValidatePath(string(*p))
}

var ErrResourceNotFound = errors.New("resource does not exist")

type Atlas struct {
	root string
}

// Creates new filesystem and creates basic dir structure
func NewAtlas(root string) (*Atlas, error) {
	atlas := &Atlas{root: filepath.Join(root, "atlas")}

	err := os.MkdirAll(root, filePerm)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(root, "atlas", "curr"), filePerm)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(root, "atlas", "tags"), filePerm)
	if err != nil {
		return nil, err
	}

	return atlas, nil
}

func (f *Atlas) Exists(path Path) bool {
	if _, err := os.Stat(path.Resolve(f)); err != nil {
		return false
	}
	return true
}

var ErrUploadToRoot = errors.New("Cannot write to root")

func (f *Atlas) Write(path Path) (io.Writer, error) {
	if err := path.Validate(); err != nil {
		return nil, err
	}

	if path == "" {
		return nil, ErrUploadToRoot
	}

	p := path.Resolve(f)

	err := os.MkdirAll(filepath.Dir(p), filePerm)
	if err != nil {
		return nil, err
	}

	f.Delete(path)
	file, err := os.Create(p)
	if err != nil {
		return nil, err
	}

	return file, nil
}

var ErrIsFolder = errors.New("Expecting file, found folder")

func (f *Atlas) Read(path Path) (io.Reader, error) {
	if err := path.Validate(); err != nil {
		return nil, err
	}

	p := path.Resolve(f)

	info, err := path.Stat(f)
	if info.IsDir() {
		return nil, ErrIsFolder
	}

	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (f *Atlas) Delete(path Path) error {
	if !f.Exists(path) {
		return ErrResourceNotFound
	}
	if err := path.Validate(); err != nil {
		return err
	}
	if err := os.RemoveAll(path.Resolve(f)); err != nil {
		return err
	}

	return nil
}

func (f *Atlas) MakeTag(name string, path Path) error {
	return nil
}

func (f *Atlas) DeleteTag(name string) error {
	return nil
}
