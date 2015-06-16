#SimpleJSON2 [![Build Status](https://travis-ci.org/xyproto/simplejson2.svg?branch=master)](https://travis-ci.org/xyproto/simplejson2) [![Build Status](https://drone.io/github.com/xyproto/simplejson2/status.png)](https://drone.io/github.com/xyproto/simplejson2/latest) [![GoDoc](https://godoc.org/github.com/xyproto/simplejson2?status.svg)](http://godoc.org/github.com/xyproto/simplejson2)

Interact with arbitrary JSON.

### Example usage

~~~go
package main

import (
	"fmt"
	"github.com/xyproto/simplejson2"
	"log"
)

func main() {
	// Some JSON
	data := []byte(`{"a":2, "b":3}`)

	// Create a new *simplejson.JSON struct
	js, err := simplejson.NewJSON(data)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the value of "a", as an int
	val, err := js.Get("a").Int()
	if err != nil {
		log.Fatal(err)
	}

	// Output the result
	fmt.Println("a is", val)
}
~~~
