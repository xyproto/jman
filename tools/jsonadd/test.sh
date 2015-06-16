#!/bin/sh
cp books.orig books.json
go run main.go books.json x '{"author": "Catniss", "book": "Yeah"}'
