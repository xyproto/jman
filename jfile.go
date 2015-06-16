package jman

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
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

// Add two byte slices together
func badd(a, b []byte) []byte {
	var buf bytes.Buffer
	buf.Write(a)
	buf.Write(b)
	return buf.Bytes()
}

// Recursively look up a given JSON path
func jsonpath(js *Node, JSONpath string) (*Node, string, error) {
	if JSONpath == "" {
		// Could not find end node
		return js, JSONpath, ErrSpecificNode
	}
	firstpart := JSONpath
	secondpart := ""
	if strings.Contains(JSONpath, ".") {
		fields := strings.SplitN(JSONpath, ".", 2)
		firstpart = fields[0]
		secondpart = fields[1]
	}
	if firstpart == "x" {
		return jsonpath(js, secondpart)
	} else if strings.Contains(firstpart, "[") && strings.Contains(firstpart, "]") {
		fields := strings.SplitN(firstpart, "[", 2)
		name := fields[0]
		if name != "x" {
			js = js.Get(name)
		}
		fields = strings.SplitN(fields[1], "]", 2)
		index, err := strconv.Atoi(fields[0])
		if err != nil {
			return js, JSONpath, errors.New("Invalid index: " + fields[0] + " (" + err.Error() + ")")
		}
		node, ok := js.GetIndex(index)
		if !ok {
			return js, JSONpath, errors.New("Could not find index: " + fields[0])
		}
		return jsonpath(node, secondpart)
	}
	name := firstpart
	if secondpart != "" {
		return js, JSONpath, errors.New("JSON path left unparsed: " + secondpart)
	}
	return js.Get(name), "", nil
}

// GetNode will find the JSON node that corresponds to the given JSON path
func (jf *JFile) GetNode(JSONpath string) (*Node, error) {
	foundnode, leftoverpath, err := jsonpath(jf.rootnode, JSONpath)
	if err != nil {
		return nil, err
	}
	if leftoverpath != "" {
		return nil, errors.New("JSON path left unparsed: " + leftoverpath)
	}
	if foundnode == nil {
		return nil, errors.New("Could not lookup: " + JSONpath)
	}
	return foundnode, nil
}

// GetString will find the string that corresponds to the given JSON path
func (jf *JFile) GetString(JSONpath string) (string, error) {
	node, err := jf.GetNode(JSONpath)
	if err != nil {
		return "", err
	}
	return node.String(), nil
}

// SetString will change the value of the key that the given JSON path points to
func (jf *JFile) SetString(JSONpath, value string) error {
	firstpart := ""
	lastpart := JSONpath
	if strings.Contains(JSONpath, ".") {
		pos := strings.LastIndex(JSONpath, ".")
		firstpart = JSONpath[:pos]
		lastpart = JSONpath[pos+1:]
	}

	node, _, err := jsonpath(jf.rootnode, firstpart)
	if (err != nil) && (err != ErrSpecificNode) {
		return err
	}

	_, hasNode := node.CheckGet(lastpart)
	if !hasNode {
		return errors.New("Index out of range? Could not set value.")
	}

	// It's weird that simplejson Set does not return an error value
	node.Set(lastpart, value)

	newdata, err := jf.rootnode.EncodePretty()
	if err != nil {
		return err
	}

	return jf.Write(newdata)
}

// Write writes the current JSON data to the file
func (jf *JFile) Write(data []byte) error {
	jf.rw.Lock()
	defer jf.rw.Unlock()
	// TODO: Add newline as well?
	return ioutil.WriteFile(jf.filename, data, 0666)
}

// AddJSON adds JSON data at the given JSON path
func (jf *JFile) AddJSON(JSONpath, JSONdata string) error {
	firstpart := ""
	lastpart := JSONpath
	if strings.Contains(JSONpath, ".") {
		pos := strings.LastIndex(JSONpath, ".")
		firstpart = JSONpath[:pos]
		lastpart = JSONpath[pos+1:]
	}

	node, _, err := jsonpath(jf.rootnode, firstpart)
	if (err != nil) && (err != ErrSpecificNode) {
		return err
	}

	_, hasNode := node.CheckGet(lastpart)
	if hasNode {
		return errors.New("The JSON path should not point to a single key when adding JSON data.")
	}

	listJSON, err := node.Encode()
	if err != nil {
		return err
	}

	fullJSON, err := jf.rootnode.Encode()
	if err != nil {
		return err
	}

	// TODO: Implement a more efficient way of adding data.
	newFullJSON := bytes.Replace(fullJSON, listJSON, badd(listJSON[:len(listJSON)-1], []byte(","+JSONdata+"]")), 1)

	js, err := New(newFullJSON)
	if err != nil {
		return err
	}

	// Update the root node
	jf.rootnode = js

	newFullJSON, err = js.EncodePretty()
	if err != nil {
		return err
	}

	return jf.Write(newFullJSON)
}

// GetAll returns the current JSON data
func (jf *JFile) GetAll() ([]byte, error) {
	return jf.rootnode.EncodePretty()
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
func AddJSON(filename, JSONpath, JSONdata string) error {
	jf, err := NewFile(filename)
	if err != nil {
		return err
	}
	return jf.AddJSON(JSONpath, JSONdata)
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
