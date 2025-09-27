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

	w, err := atlas.Write("file.txt")
	assert.NoError(t, err)
	fmt.Fprint(w, "hi")

	r, err := atlas.Read("file.txt")
	assert.NoError(t, err)

	val, err := io.ReadAll(r)
	assert.Exactly(t, "hi", string(val))

	l, err := atlas.List("/")
	assert.NoError(t, err)

	assert.Exactly(t, []string{"file.txt"}, l)

	_, err = atlas.Write("file2.txt")
	assert.NoError(t, err)

	l, err = atlas.List("/")
	assert.NoError(t, err)

	assert.ElementsMatch(t, []string{"file2.txt", "file.txt"}, l)

	_, err = atlas.Write("folder/file.txt")
	assert.NoError(t, err)

	tree, err := atlas.Tree("/")
	assert.NoError(t, err)

	assert.ElementsMatch(t, []string{"file.txt", "file2.txt", "folder/file.txt"}, tree)

	assert.True(t, atlas.Exists("file.txt"))
	assert.NoError(t, atlas.Delete("file.txt"))
	assert.False(t, atlas.Exists("file.txt"))
}
