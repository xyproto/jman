// Package lookup provides a way to search and manipulate JSON files using simple JSON path expressions
package simplejson

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrSpecificNode = errors.New("Could not find a specific node that matched the given path")
)

// JSONFile represents a JSON file and contains the filename and root node
type JSONFile struct {
	filename string
	rootnode *Node
	rw       *sync.RWMutex
}

// NewJSONFile will read the given filename and return a JSONFile struct
func NewJSONFile(filename string) (*JSONFile, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	js, err := NewJSON(data)
	if err != nil {
		return nil, err
	}
	rw := &sync.RWMutex{}
	return &JSONFile{filename, js, rw}, nil
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
		return jsonpath(js.GetIndex(index), secondpart)
	}
	name := firstpart
	if secondpart != "" {
		return js, JSONpath, errors.New("JSON path left unparsed: " + secondpart)
	}
	return js.Get(name), "", nil
}

// Lookup will find the JSON node that corresponds to the given JSON path
func (jf *JSONFile) Lookup(JSONpath string) (*Node, error) {
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

// LookupString will find the string that corresponds to the given JSON Path
func (jf *JSONFile) LookupString(JSONpath string) (string, error) {
	foundnode, err := jf.Lookup(JSONpath)
	if err != nil {
		return "", err
	}
	result, err := foundnode.String()
	if err != nil {
		s := fmt.Sprint(foundnode)
		return s, fmt.Errorf("Result was not a string: %v", s)
	}
	return result, nil
}

func (jf *JSONFile) SetString(JSONpath, value string) error {
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

func (jf *JSONFile) Write(data []byte) error {
	jf.rw.Lock()
	defer jf.rw.Unlock()
	// TODO: Add newline as well?
	return ioutil.WriteFile(jf.filename, data, 0666)
}

func (jf *JSONFile) AddJSON(JSONpath, JSONdata string) error {
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

	// TODO: Fork simplejson for a better way of adding data!
	newFullJSON := bytes.Replace(fullJSON, listJSON, badd(listJSON[:len(listJSON)-1], []byte(","+JSONdata+"]")), 1)

	js, err := NewJSON(newFullJSON)
	if err != nil {
		return err
	}

	newFullJSON, err = js.EncodePretty()
	if err != nil {
		return err
	}

	return jf.Write(newFullJSON)
}

func SetString(filename, JSONpath, value string) error {
	jf, err := NewJSONFile(filename)
	if err != nil {
		return err
	}
	return jf.SetString(JSONpath, value)
}

func AddJSON(filename, JSONpath, JSONdata string) error {
	jf, err := NewJSONFile(filename)
	if err != nil {
		return err
	}
	return jf.AddJSON(JSONpath, JSONdata)
}

// JSONString will find the string that corresponds to the given JSON Path,
// given a filename and a simple JSON path expression.
func LookupString(filename, JSONpath string) (string, error) {
	jf, err := NewJSONFile(filename)
	if err != nil {
		return "", err
	}
	return jf.LookupString(JSONpath)
}
