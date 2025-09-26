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
}
