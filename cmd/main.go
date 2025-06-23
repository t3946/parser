package main

import (
	"flag"
	"log"
	"math/rand"
	browserCtl "parser/services/browserctl"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	queryPtr := flag.String("query", "", "Search query (e.g., 'купить машину')")
	lrPtr := flag.String("lr", "213", "Region code (e.g., 213 for Moscow)")
	flag.Parse()

	if *queryPtr == "" {
		log.Fatal("query is required")
	}

	browserCtl.Search(*queryPtr, *lrPtr)
}
