package main

import (
	"flag"
	"fmt"
	"github.com/xyproto/lookup"
	"log"
	"os"
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 3 {
		fmt.Println("syntax: insert [filename] [JSON path] [value]")
		fmt.Println("example: insert books.json x[1].author Catniss")
		os.Exit(1)
	}

	filename := flag.Args()[0]
	JSONpath := flag.Args()[1]
	value := flag.Args()[2]

	err := lookup.JSONSet(filename, JSONpath, value)
	if err != nil {
		log.Fatal(err)
	}
}
