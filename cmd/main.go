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
	test := "1 keywords"
	dataJson := storage.ReadFile("test/" + test + ".json")
	var kw []string
	json.Unmarshal([]byte(dataJson), &kw)
	items, stats := searchYandex.ParseKeywordsList(kw, "46")
	storage.WriteFile(test+"/result.json", items)
	storage.WriteFile(test+"/stats.json", stats)
}
