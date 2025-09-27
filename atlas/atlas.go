package atlas

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const dirPerm = 0777

type Atlas struct {
	root string
}

// Creates new filesystem and creates basic dir structure
func NewAtlas(root string) (*Atlas, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	atlas := &Atlas{root: filepath.Join(root, "atlas")}

	err = os.MkdirAll(root, dirPerm)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(root, "atlas", "fs"), dirPerm)
	if err != nil {
		return nil, err
	}

	return atlas, nil
}

func (a *Atlas) ResolvePath(path string) (string, error) {
	path = filepath.Join(a.root, "fs", path)
	if err := a.ValidatePath(path); err != nil {
		return "", err
	}

	return path, nil
}

func (a *Atlas) ValidatePath(path string) error {
	clean, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(a.root, clean)
	if err != nil {
		return fmt.Errorf("cannot evaluate path: %w", err)
	}

	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path escapes base directory: %s", path)
	}

	return nil
}

func (a *Atlas) Exists(path string) bool {
	path, err := a.ResolvePath(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	if err != nil {
		return false
	}

	return true
}

func (a *Atlas) List(path string) ([]string, error) {
	path, err := a.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, ErrNotFolder
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	list := make([]string, len(entries))
	for i, entry := range entries {
		list[i] = entry.Name()
	}

	return list, nil
}

func (a *Atlas) tree(path string) ([]string, error) {
	fmt.Println("path: ", path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, ErrNotFolder
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			subpath := entry.Name()
			subtree, err := a.tree(filepath.Join(path, subpath))
			if err != nil {
				return nil, err
			}
			list = append(list, subtree...)
		} else {
			list = append(list, filepath.Join(path, entry.Name()))
		}
	}

	return list, nil
}

func (a *Atlas) Tree(path string) ([]string, error) {
	path, err := a.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	tree, err := a.tree(path)
	if err != nil {
		return nil, err
	}

	for i := range tree {
		tree[i], err = filepath.Rel(path, tree[i])
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func (a *Atlas) Write(path string) (io.Writer, error) {
	path, err := a.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	if path == a.root {
		return nil, ErrUploadToRoot
	}

	err = os.MkdirAll(filepath.Dir(path), dirPerm)
	if err != nil {
		return nil, err
	}

	a.Delete(path)
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (a *Atlas) Read(path string) (io.Reader, error) {
	path, err := a.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, ErrResourceNotFound
	}
	if info.IsDir() {
		return nil, ErrNotFile
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (a *Atlas) Delete(path string) error {
	path, err := a.ResolvePath(path)
	if err != nil {
		return err
	}

	if !a.Exists(path) {
		return ErrResourceNotFound
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}
