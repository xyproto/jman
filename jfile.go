package jman

import (
	"errors"
	"io/ioutil"
	"sync"
)

var (
	// ErrSpecificNode is for when retrieving a node does not return a specific key/value, but perhaps a map
	ErrSpecificNode = errors.New("Could not find a specific node that matched the given path")
)

// JFile represents a JSON file and contains the filename and root node
type JFile struct {
	filename string
	rootnode *Node
	rw       *sync.RWMutex
}

// NewFile will read the given filename and return a JFile struct
func NewFile(filename string) (*JFile, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	js, err := New(data)
	if err != nil {
		return nil, err
	}
	rw := &sync.RWMutex{}
	return &JFile{filename, js, rw}, nil
}

// GetNode tries to find the JSON node that corresponds to the given JSON path
func (jf *JFile) GetNode(JSONpath string) (*Node, error) {
	node, _, err := jf.rootnode.GetNodes(JSONpath)
	return node, err
}

// GetString tries to find the string that corresponds to the given JSON path
func (jf *JFile) GetString(JSONpath string) (string, error) {
	node, err := jf.GetNode(JSONpath)
	if err != nil {
		return "", err
	}
	return node.String(), nil
}

// SetString will change the value of the key that the given JSON path points to
func (jf *JFile) SetString(JSONpath, value string) error {
	_, parentNode, err := jf.rootnode.GetNodes(JSONpath)
	if err != nil {
		return err
	}
	m, ok := parentNode.CheckMap()
	if !ok {
		return errors.New("Parent is not a map: " + JSONpath)
	}

	// Set the string
	m[lastpart(JSONpath)] = value

	newdata, err := jf.rootnode.PrettyJSON()
	if err != nil {
		return err
	}

	return jf.Write(newdata)
}

// Write writes the current JSON data to the file
func (jf *JFile) Write(data []byte) error {
	jf.rw.Lock()
	defer jf.rw.Unlock()
	return ioutil.WriteFile(jf.filename, data, 0666)
}

// AddJSON adds JSON data at the given JSON path
func (jf *JFile) AddJSON(JSONpath string, JSONdata []byte, pretty bool) error {
	jf.rootnode.AddJSON(JSONpath, JSONdata)
	var (
		data []byte
		err  error
	)
	if pretty {
		data, err = jf.rootnode.PrettyJSON()
	} else {
		data, err = jf.rootnode.JSON()
	}
	if err != nil {
		return err
	}
	return jf.Write(data)
}

// JSON returns the current JSON data
func (jf *JFile) JSON() ([]byte, error) {
	return jf.rootnode.PrettyJSON()
}

// SetString sets a value to the given JSON file at the given JSON path
func SetString(filename, JSONpath, value string) error {
	jf, err := NewFile(filename)
	if err != nil {
		return err
	}
	return jf.SetString(JSONpath, value)
}

// AddJSON adds JSON data to the given JSON file at the given JSON path
func AddJSON(filename, JSONpath string, JSONdata []byte, pretty bool) error {
	jf, err := NewFile(filename)
	if err != nil {
		return err
	}
	return jf.AddJSON(JSONpath, JSONdata, pretty)
}

// GetString will find the string that corresponds to the given JSON Path,
// given a filename and a simple JSON path expression.
func GetString(filename, JSONpath string) (string, error) {
	jf, err := NewFile(filename)
	if err != nil {
		return "", err
	}
	return jf.GetString(JSONpath)
}
