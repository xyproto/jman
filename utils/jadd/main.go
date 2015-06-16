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

	if len(flag.Args()) != 3 {
		fmt.Println("syntax: add [filename] [JSON path] [JSON data]")
		fmt.Println("example: add books.json x '{\"author\": \"Catniss\", \"book\": \"Yeah\"}'")
		os.Exit(1)
	}

	filename := flag.Args()[0]
	JSONpath := flag.Args()[1]
	JSONdata := flag.Args()[2]

	err := jman.AddJSON(filename, JSONpath, JSONdata)
	if err != nil {
		log.Fatal(err)
	}
}
