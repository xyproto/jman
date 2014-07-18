// Package jman provides a way to search and manipulate JSON documents
package jman

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"reflect"
	"strconv"
)

// Stable API within the same major version number
const Version = 1.0

// Node can be a JSON document, or a part of a JSON document
type (
	Node struct {
		data interface{}
	}
	NodeSlice []*Node
	DuckSlice []interface{}
	NodeMap   map[string]*Node
	DuckMap   map[string]interface{}
)

var NilNode = &Node{nil}

// New returns a pointer to a new `Node` object
// after unmarshaling `body` bytes
func New(body []byte) (*Node, error) {
	j := new(Node)
	err := j.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// NewNode returns a pointer to a new, empty `Node` object
func NewNode() *Node {
	return &Node{
		data: make(DuckMap),
	}
}

// Interface returns the underlying data
func (j *Node) Interface() interface{} {
	return j.data
}

// Encode returns its marshaled data as `[]byte`
func (j *Node) Encode() ([]byte, error) {
	return j.MarshalJSON()
}

// EncodePretty returns its marshaled data as `[]byte` with indentation
func (j *Node) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

// MarshalJSON implements the json.Marshaler interface
func (j *Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

// Set modifies `Node` map by `key` and `value`
// Useful for changing single key/value in a `Node` object easily.
func (j *Node) Set(key string, val interface{}) {
	m, ok := j.CheckMap()
	if !ok {
		return
	}
	m[key] = val
}

// SetPath modifies `Node`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value.
func (j *Node) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		j.data = val
		return
	}

	// in order to insert our branch, we need DuckMap
	if _, ok := (j.data).(DuckMap); !ok {
		// have to replace with something suitable
		j.data = make(DuckMap)
	}
	curr := j.data.(DuckMap)

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(DuckMap)
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(DuckMap); !ok {
			// have to replace with something suitable
			n := make(DuckMap)
			curr[b] = n
		}

		curr = curr[b].(DuckMap)
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

// Del modifies `Node` map by deleting `key` if it is present.
func (j *Node) Del(key string) {
	m, ok := j.CheckMap()
	if !ok {
		return
	}
	delete(m, key)
}

// GetKey returns a pointer to a new `Node` object
// for `key` in its `map` representation
// and a bool identifying success or failure
func (j *Node) GetKey(key string) (*Node, bool) {
	m, ok := j.CheckMap()
	if ok {
		if val, ok := m[key]; ok {
			return &Node{val}, true
		}
	}
	return nil, false
}

// GetIndex returns a pointer to a new `Node` object
// for `index` in its `array` representation
// and a bool identifying success or failure
func (j *Node) GetIndex(index int) (*Node, bool) {
	a, ok := j.CheckSlice()
	if ok {
		if len(a) > index {
			return &Node{a[index]}, true
		}
	}
	return nil, false
}

// Get searches for the item as specified by the branch
// within a nested Json and returns a new Json pointer
// the pointer is always a valid Json, allowing for chained operations
//
//   newJs := js.Get("top_level", "entries", 3, "dict")
func (j *Node) Get(branch ...interface{}) *Node {
	jin, ok := j.CheckGet(branch...)
	if ok {
		return jin
	} else {
		return NilNode
	}
}

// CheckGet is like Get, except it also returns a bool
// indicating whenever the branch was found or not
// the Json pointer may be nil
//
//   newJs, ok := js.Get("top_level", "entries", 3, "dict")
func (j *Node) CheckGet(branch ...interface{}) (*Node, bool) {
	jin := j
	var ok bool
	for _, p := range branch {
		switch p.(type) {
		case string:
			jin, ok = jin.GetKey(p.(string))
		case int:
			jin, ok = jin.GetIndex(p.(int))
		default:
			ok = false
		}
		if !ok {
			return nil, false
		}
	}
	return jin, true
}

// ChechNodeMap returns a copy of a Json map, but with values as Jsons
func (j *Node) CheckNodeMap() (NodeMap, bool) {
	m, ok := j.CheckMap()
	if !ok {
		return nil, false
	}
	jm := make(NodeMap)
	for key, val := range m {
		jm[key] = &Node{val}
	}
	return jm, true
}

// CheckNodeSlice returns a copy of an array, but with each value as a Json
func (j *Node) CheckNodeSlice() ([]*Node, bool) {
	a, ok := j.CheckSlice()
	if !ok {
		return nil, false
	}
	ja := make([]*Node, len(a))
	for key, val := range a {
		ja[key] = &Node{val}
	}
	return ja, true
}

// CheckMap type asserts to `map`
func (j *Node) CheckMap() (DuckMap, bool) {
	if m, ok := (j.data).(DuckMap); ok {
		return m, true
	}
	return nil, false
}

// CheckSlice type asserts to an `array`
func (j *Node) CheckSlice() (DuckSlice, bool) {
	if a, ok := (j.data).(DuckSlice); ok {
		return a, true
	}
	return nil, false
}

// CheckBool type asserts to `bool`
func (j *Node) CheckBool() (bool, bool) {
	if s, ok := (j.data).(bool); ok {
		return s, true
	}
	return false, false
}

// CheckString type asserts to `string`
func (j *Node) CheckString() (string, bool) {
	if s, ok := (j.data).(string); ok {
		return s, true
	}
	return "", false
}

// NodeSlice guarantees the return of a `[]interface{}` (with optional default)
func (j *Node) NodeSlice(args ...NodeSlice) NodeSlice {
	var def NodeSlice

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("NodeSlice() received too many arguments %d", len(args))
	}

	a, ok := j.CheckNodeSlice()
	if ok {
		return a
	}

	return def
}

// NodeMap guarantees the return of a `map[string]*Node` (with optional default)
func (j *Node) NodeMap(args ...NodeMap) NodeMap {
	var def NodeMap

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("NodeMap() received too many arguments %d", len(args))
	}

	if a, ok := j.CheckNodeMap(); ok {
		return a
	}

	return def
}

// Slice guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, v := range js.Get("results").Slice() {
//			fmt.Println(i, v)
//		}
func (j *Node) Slice(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Slice() received too many arguments %d", len(args))
	}

	a, ok := j.CheckSlice()
	if ok {
		return a
	}

	return def
}

// Map guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//		for k, v := range js.Get("dictionary").Map() {
//			fmt.Println(k, v)
//		}
func (j *Node) Map(args ...DuckMap) DuckMap {
	var def DuckMap

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Map() received too many arguments %d", len(args))
	}

	a, ok := j.CheckMap()
	if ok {
		return a
	}

	return def
}

// String guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.Get("param1").String(), js.Get("optional_param").String("my_default"))
func (j *Node) String(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("String() received too many arguments %d", len(args))
	}

	s, ok := j.CheckString()
	if ok {
		return s
	}

	return def
}

// Int guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.Get("param1").Int(), js.Get("optional_param").Int(5150))
func (j *Node) Int(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Int() received too many arguments %d", len(args))
	}

	i, ok := j.CheckInt()
	if ok {
		return i
	}

	return def
}

// Float64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Get("param1").Float64(), js.Get("optional_param").Float64(5.150))
func (j *Node) Float64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Float64() received too many arguments %d", len(args))
	}

	f, ok := j.CheckFloat64()
	if ok {
		return f
	}

	return def
}

// Bool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.Get("param1").Bool(), js.Get("optional_param").Bool(true))
func (j *Node) Bool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Bool() received too many arguments %d", len(args))
	}

	b, ok := j.CheckBool()
	if ok {
		return b
	}

	return def
}

// Int64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Get("param1").Int64(), js.Get("optional_param").Int64(5150))
func (j *Node) Int64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Int64() received too many arguments %d", len(args))
	}

	i, ok := j.CheckInt64()
	if ok {
		return i
	}

	return def
}

// Uint64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Get("param1").Uint64(), js.Get("optional_param").Uint64(5150))
func (j *Node) Uint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Uint64() received too many arguments %d", len(args))
	}

	i, ok := j.CheckUint64()
	if ok {
		return i
	}

	return def
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (j *Node) UnmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(p))
	dec.UseNumber()
	return dec.Decode(&j.data)
}

// NewFromReader returns a *Node by decoding from an io.Reader
func NewFromReader(r io.Reader) (*Node, error) {
	j := new(Node)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	err := dec.Decode(&j.data)
	return j, err
}

// CheckFloat64 coerces into a float64
func (j *Node) CheckFloat64() (float64, bool) {
	switch j.data.(type) {
	case json.Number:
		nr, err := j.data.(json.Number).Float64()
		return nr, err == nil
	case float32, float64:
		return reflect.ValueOf(j.data).Float(), true
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(j.data).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(j.data).Uint()), true
	}
	return 0, false
}

// CheckInt coerces into an int
func (j *Node) CheckInt() (int, bool) {
	switch j.data.(type) {
	case json.Number:
		nr, err := j.data.(json.Number).Int64()
		return int(nr), err == nil
	case float32, float64:
		return int(reflect.ValueOf(j.data).Float()), true
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(j.data).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(j.data).Uint()), true
	}
	return 0, false
}

// CheckInt64 coerces into an int64
func (j *Node) CheckInt64() (int64, bool) {
	switch j.data.(type) {
	case json.Number:
		nr, err := j.data.(json.Number).Int64()
		return nr, err == nil
	case float32, float64:
		return int64(reflect.ValueOf(j.data).Float()), true
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(j.data).Int(), true
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(j.data).Uint()), true
	}
	return 0, false
}

// CheckUint64 coerces into an uint64
func (j *Node) CheckUint64() (uint64, bool) {
	switch j.data.(type) {
	case json.Number:
		nr, err := strconv.ParseUint(j.data.(json.Number).String(), 10, 64)
		return nr, err == nil
	case float32, float64:
		return uint64(reflect.ValueOf(j.data).Float()), true
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(j.data).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(j.data).Uint(), true
	}
	return 0, false
}
