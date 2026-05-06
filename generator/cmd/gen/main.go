package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: gen <path/to/joltc.h>")
	}
	log.Printf("generator scaffold; input=%s (parsing wired in next tasks)", os.Args[1])
}
