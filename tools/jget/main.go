package main

import (
	"flag"
	"fmt"
	"github.com/xyproto/jman"
	"log"
	"os"
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Println("syntax: lookup [filename] [JSON path]")
		fmt.Println("example: lookup books.json x[1].author")
		os.Exit(1)
	}

	filename := flag.Args()[0]
	JSONpath := flag.Args()[1]

	foundString, err := jman.GetString(filename, JSONpath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(foundString)
}
