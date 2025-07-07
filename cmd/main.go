package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"parser/services/searchYandex"
	"parser/services/storage"
)

func init() {
	godotenv.Load()
}

func main() {
	dataJson := storage.ReadFile("test/100 keywords.json")
	var kw []string
	json.Unmarshal([]byte(dataJson), &kw)
	items, stats := searchYandex.ParseKeywordsList(kw[0:3], "46")
	storage.WriteFile("load-kw-test/result.json", items)
	storage.WriteFile("load-kw-test/stats.json", stats)
}
