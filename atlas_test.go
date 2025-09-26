package main

import (
	"fmt"
	"testing"
)

func TestTree(t *testing.T) {
	atlas, err := NewAtlas(".")
	if err != nil {
		t.Error(err)
	}

	w, err := atlas.Write("file.txt")
	if err != nil {
		t.Error(err)
	}
	fmt.Fprint(w, "hi")
}
