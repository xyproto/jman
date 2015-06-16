#Simple JSON manager [![Build Status](https://travis-ci.org/xyproto/jman.svg?branch=master)](https://travis-ci.org/xyproto/jman) [![GoDoc](https://godoc.org/github.com/xyproto/jman?status.svg)](http://godoc.org/github.com/xyproto/jman)

Interact with arbitrary JSON.

### Example usage

~~~go
package main

import (
	"fmt"
	"github.com/xyproto/jman"
	"log"
)

func main() {
	// Some JSON
	data := []byte(`{"a":2, "b":3}`)

	// Create a new *jman.Node
	js, err := jman.New(data)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the value of "a", as an int
	val := js.Get("a").Int()

	// Output the result
	fmt.Println("a is", val)
}
~~~

### JSON paths

Several of the `JFile` methods takes a simple JSON path expression, like `x.books[1].author`. Only simple expressions using `x` for the root node, names and integer indexes are supported. For more advanced JSON path expressions, see [this blog post](http://goessner.net/articles/JsonPath/).

The `SetBranch` method for the `Node` struct also provides a way of accessing JSON nodes, where the JSON names are supplied as a slice of strings.

### Requirements

* go >= 1.2
