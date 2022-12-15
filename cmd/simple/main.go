package main

import (
	"fmt"
	"github.com/xyproto/jpath"
	"log"
)

func main() {
	// Some JSON
	data := []byte(`{"a":2, "b":3, "people":{"names": ["Bob", "Alice"]}}`)

	// Create a new *jpath.Node
	document, err := jpath.New(data)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the value of "a", as an int
	val := document.Get("a").Int()
	fmt.Println("a is", val)

	// Retrieve the first name, using a path expression
	name := document.GetNode(".people.names[0]").String()
	fmt.Println("The name is", name)
}
