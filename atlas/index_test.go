package atlas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	index, err := OpenIndex(":memory:")
	assert.Nil(t, err, "failed to open index")

	assert.Nil(t, index.Init(), "failed to initialise index")

	// Check Insert
	assert.Nil(t, index.Put("/file.txt", 0, "hash"), "failed to initialise index")

	// Check GetPath
	row, err := index.GetPath("/file.txt")
	assert.Nil(t, err)

	assert.Equal(t, IndexRow{
		"/file.txt", 0, "hash",
	}, row)

	// Check GetPath on invalid path
	row, err = index.GetPath("/file1.txt")
	assert.Error(t, err)

	assert.Nil(t, index.Put("/file.txt", 1, "hash"), "failed to initialise index")

	// Check Insert on existing path
	row, err = index.GetPath("/file.txt")
	assert.Nil(t, err)

	assert.Equal(t, IndexRow{
		"/file.txt", 1, "hash",
	}, row)

	// Check Delete
	assert.Nil(t, index.DeletePath("/file.txt"))
	row, err = index.GetPath("/file.txt")
	assert.Error(t, err)
}
