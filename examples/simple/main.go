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
