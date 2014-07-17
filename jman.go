// Package jman provides a way to search and manipulate JSON documents
package jman

import (
	"bytes"
	"encoding/json"
	"errors"
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
	NodeMap   map[string]*Node
	AnyMap    map[string]interface{}
)

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
		data: make(AnyMap),
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

// Set modifies `Node` map by `key` and `value`.
// Useful for changing single key/value in a `Node` object easily.
func (j *Node) Set(key string, val interface{}) {
	m, err := j.Map()
	if err != nil {
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

	// in order to insert our branch, we need AnyMap
	if _, ok := (j.data).(AnyMap); !ok {
		// have to replace with something suitable
		j.data = make(AnyMap)
	}
	curr := j.data.(AnyMap)

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(AnyMap)
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(AnyMap); !ok {
			// have to replace with something suitable
			n := make(AnyMap)
			curr[b] = n
		}

		curr = curr[b].(AnyMap)
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

// Del modifies `Node` map by deleting `key` if it is present.
func (j *Node) Del(key string) {
	m, err := j.Map()
	if err != nil {
		return
	}
	delete(m, key)
}

// GetKey returns a pointer to a new `Node` object
// for `key` in its `map` representation
// and a bool identifying success or failure
func (j *Node) GetKey(key string) (*Node, bool) {
	m, err := j.Map()
	if err == nil {
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
	a, err := j.Array()
	if err == nil {
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
		return &Node{nil}
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

// JsonMap returns a copy of a Json map, but with values as Nodes
func (j *Node) JsonMap() (NodeMap, error) {
	m, err := j.Map()
	if err != nil {
		return nil, err
	}
	jm := make(NodeMap)
	for key, val := range m {
		jm[key] = &Node{val}
	}
	return jm, nil
}

// JsonArray returns a copy of an array, but with each value as a Node
func (j *Node) JsonArray() ([]*Node, error) {
	a, err := j.Array()
	if err != nil {
		return nil, err
	}
	ja := make([]*Node, len(a))
	for key, val := range a {
		ja[key] = &Node{val}
	}
	return ja, nil
}

// Map type asserts to `map`
func (j *Node) Map() (AnyMap, error) {
	if m, ok := (j.data).(AnyMap); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// Array type asserts to an `array`
func (j *Node) Array() ([]interface{}, error) {
	if a, ok := (j.data).([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// Bool type asserts to `bool`
func (j *Node) Bool() (bool, error) {
	if s, ok := (j.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

// String type asserts to `string`
func (j *Node) String() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// Bytes type asserts to `[]byte`
func (j *Node) Bytes() ([]byte, error) {
	if s, ok := (j.data).(string); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

// MustJsonArray guarantees the return of a `[]interface{}` (with optional default)
func (j *Node) MustJsonArray(args ...NodeSlice) NodeSlice {
	var def NodeSlice

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustJsonArray() received too many arguments %d", len(args))
	}

	a, err := j.JsonArray()
	if err == nil {
		return a
	}

	return def
}

// MustJsonMap guarantees the return of a `map[string]interface{}` (with optional default)
func (j *Node) MustJsonMap(args ...NodeMap) NodeMap {
	var def NodeMap

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustJsonMap() received too many arguments %d", len(args))
	}

	a, err := j.JsonMap()
	if err == nil {
		return a
	}

	return def
}

// MustArray guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, v := range js.Get("results").MustArray() {
//			fmt.Println(i, v)
//		}
func (j *Node) MustArray(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustArray() received too many arguments %d", len(args))
	}

	a, err := j.Array()
	if err == nil {
		return a
	}

	return def
}

// MustMap guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//		for k, v := range js.Get("dictionary").MustMap() {
//			fmt.Println(k, v)
//		}
func (j *Node) MustMap(args ...AnyMap) AnyMap {
	var def AnyMap

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	a, err := j.Map()
	if err == nil {
		return a
	}

	return def
}

// MustString guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.Get("param1").MustString(), js.Get("optional_param").MustString("my_default"))
func (j *Node) MustString(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustString() received too many arguments %d", len(args))
	}

	s, err := j.String()
	if err == nil {
		return s
	}

	return def
}

// MustInt guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.Get("param1").MustInt(), js.Get("optional_param").MustInt(5150))
func (j *Node) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt() received too many arguments %d", len(args))
	}

	i, err := j.Int()
	if err == nil {
		return i
	}

	return def
}

// MustFloat64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Get("param1").MustFloat64(), js.Get("optional_param").MustFloat64(5.150))
func (j *Node) MustFloat64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustFloat64() received too many arguments %d", len(args))
	}

	f, err := j.Float64()
	if err == nil {
		return f
	}

	return def
}

// MustBool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (j *Node) MustBool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustBool() received too many arguments %d", len(args))
	}

	b, err := j.Bool()
	if err == nil {
		return b
	}

	return def
}

// MustInt64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Get("param1").MustInt64(), js.Get("optional_param").MustInt64(5150))
func (j *Node) MustInt64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt64() received too many arguments %d", len(args))
	}

	i, err := j.Int64()
	if err == nil {
		return i
	}

	return def
}

// MustUint64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Get("param1").MustUint64(), js.Get("optional_param").MustUint64(5150))
func (j *Node) MustUint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustUint64() received too many arguments %d", len(args))
	}

	i, err := j.Uint64()
	if err == nil {
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

// Float64 coerces into a float64
func (j *Node) Float64() (float64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(j.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int coerces into an int
func (j *Node) Int() (int, error) {
	switch j.data.(type) {
	case json.Number:
		i, err := j.data.(json.Number).Int64()
		return int(i), err
	case float32, float64:
		return int(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int64 coerces into an int64
func (j *Node) Int64() (int64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(j.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Uint64 coerces into an uint64
func (j *Node) Uint64() (uint64, error) {
	switch j.data.(type) {
	case json.Number:
		return strconv.ParseUint(j.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(j.data).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}
