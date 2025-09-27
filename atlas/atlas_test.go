package atlas

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	folder, err := os.MkdirTemp("", "atlas_*")
	defer os.RemoveAll(folder)
	assert.NoError(t, err)

	atlas, err := NewAtlas(folder)
	assert.NoError(t, err)

	file, err := atlas.FSPath("file.txt")
	assert.NoError(t, err)
	w, err := atlas.Write(file)
	assert.NoError(t, err)
	fmt.Fprint(w, "hi")

	file, err = atlas.FSPath("file.txt")
	assert.NoError(t, err)
	r, err := atlas.Read(file)
	assert.NoError(t, err)

	val, err := io.ReadAll(r)
	assert.Exactly(t, "hi", string(val))

	root, err := atlas.FSPath("/")
	assert.NoError(t, err)
	l, err := atlas.List(root)
	assert.NoError(t, err)

	assert.Exactly(t, []string{"file.txt"}, l)

	file2, err := atlas.FSPath("file2.txt")
	assert.NoError(t, err)
	_, err = atlas.Write(file2)
	assert.NoError(t, err)

	l, err = atlas.List(root)
	assert.NoError(t, err)

	assert.ElementsMatch(t, []string{"file2.txt", "file.txt"}, l)

	subfile, err := atlas.FSPath("folder/file.txt")
	assert.NoError(t, err)
	_, err = atlas.Write(subfile)
	assert.NoError(t, err)

	tree, err := atlas.Tree(root)
	assert.NoError(t, err)

	assert.ElementsMatch(t, []string{"file.txt", "file2.txt", "folder/file.txt"}, tree)

	assert.NoError(t, atlas.Delete(file))
	file, err = atlas.FSPath("file.txt")
	assert.NoError(t, err)
	assert.Nil(t, file.info)
}
