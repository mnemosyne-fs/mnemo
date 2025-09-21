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
var InvalidTagErr = errors.New("tag is not valid")

func ValidateTag(tag string) error {
	if !tag_validate.Match([]byte(tag)) {
		return InvalidTagErr
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

type Path struct {
	tag    string
	path   string
	owners []string
}

func CurrPath(path string) Path {
	return Path{
		tag:  "",
		path: path,
	}
}

func TagPath(tag, path string) Path {
	return Path{
		tag:  tag,
		path: path,
	}
}

func (p *Path) Resolve(atlas *Atlas) string {
	if p.tag == "" {
		return filepath.Join(atlas.root, "curr", p.path)
	}
	return filepath.Join(atlas.root, "tag", p.tag, p.path)
}

func (p *Path) Stat(atlas *Atlas) (os.FileInfo, error) {
	info, err := os.Stat(p.Resolve(atlas))
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (p *Path) Validate() error {
	var tagerr error = nil
	if p.tag != "" {
		tagerr = ValidateTag(p.tag)
	}
	patherr := ValidatePath(p.path)

	return errors.Join(tagerr, patherr)
}

var ResourceNotFoundErr = errors.New("resource does not exist")

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

func (f *Atlas) Upload(r io.Reader, path Path) error {
	if err := path.Validate(); err != nil {
		return err
	}

	if path.path == "" {
		return fmt.Errorf("Can't upload to root")
	}

	p := path.Resolve(f)

	err := os.MkdirAll(filepath.Dir(p), filePerm)
	if err != nil {
		return err
	}

	f.Delete(path)
	file, err := os.Create(p)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, r)
	if err != nil {
		os.Remove(p)
		return err
	}

	return nil
}

func (f *Atlas) Delete(path Path) error {
	if !f.Exists(path) {
		return ResourceNotFoundErr
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
