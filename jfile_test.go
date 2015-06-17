package jman

import (
	"github.com/bmizerany/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestAddFile(t *testing.T) {
	var err error
	someJSON := []byte(`{"x":"2", "y":"3"}`)
	documentJSON := []byte(`[{"x":"7", "y":"15"}]`)
	finalJSON := []byte(`[{"x":"7","y":"15"},{"x":"2","y":"3"}]`)
	tmpfile := "/tmp/___jman.json"
	err = ioutil.WriteFile(tmpfile, documentJSON, 0666)
	assert.Equal(t, nil, err)
	defer os.Remove(tmpfile)

	err = AddJSON(tmpfile, "x", someJSON, false)
	assert.Equal(t, nil, err)

	fileData, err := ioutil.ReadFile(tmpfile)
	assert.Equal(t, nil, err)

	assert.Equal(t, string(fileData), string(finalJSON))
}
