package main

import (
	"flag"
	"log"
	"math/rand"
	"parser/services/searchYandex"
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

	searchYandex.Search(*queryPtr, *lrPtr)
}
