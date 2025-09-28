package atlas

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const dirPerm = 0777

type FSPath struct {
	path string      // Stored as absolute path on host system
	info os.FileInfo // If nil, then file does not currently exist
}

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

func (a *Atlas) FSPath(path string) (FSPath, error) {
	path = filepath.Join(a.root, "fs", path)
	abs, err := filepath.Abs(path)
	if err != nil {
		return FSPath{}, err
	}

	rel, err := filepath.Rel(a.root, abs)
	if err != nil {
		return FSPath{}, fmt.Errorf("cannot evaluate path: %w", err)
	}

	if strings.HasPrefix(rel, "..") {
		return FSPath{}, fmt.Errorf("path escapes base directory: %s", path)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return FSPath{path, nil}, nil
	}

	return FSPath{path, info}, nil
}

func (a *Atlas) List(path FSPath) ([]string, error) {
	if !path.info.IsDir() {
		return nil, ErrNotFolder
	}

	entries, err := os.ReadDir(path.path)
	if err != nil {
		return nil, err
	}

	list := make([]string, len(entries))
	for i, entry := range entries {
		list[i] = entry.Name()
	}

	return list, nil
}

func (a *Atlas) tree_helper(path string) ([]string, error) {
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
			subtree, err := a.tree_helper(filepath.Join(path, subpath))
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

func (a *Atlas) Tree(path FSPath) ([]string, error) {
	tree, err := a.tree_helper(path.path)
	if err != nil {
		return nil, err
	}

	for i := range tree {
		tree[i], err = filepath.Rel(path.path, tree[i])
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}

func (a *Atlas) Writer(path FSPath) (io.Writer, error) {
	if path.path == a.root {
		return nil, ErrUploadToRoot
	}

	err := os.MkdirAll(filepath.Dir(path.path), dirPerm)
	if err != nil {
		return nil, err
	}

	a.Delete(path)
	file, err := os.Create(path.path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (a *Atlas) Reader(path FSPath) (io.Reader, error) {
	if path.info.IsDir() {
		return nil, ErrNotFile
	}

	file, err := os.Open(path.path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (a *Atlas) Delete(path FSPath) error {
	if err := os.RemoveAll(path.path); err != nil {
		return err
	}

	return nil
}
