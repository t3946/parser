package main

import (
	"encoding/json"
	"fmt"
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
	kwNumber := 20
	items, stats := searchYandex.ParseKeywordsListWithChromeDP(kw[0:kwNumber], "46")
	dir := fmt.Sprintf("load-kw-test-%v", kwNumber)
	storage.WriteFile(dir+"/result.json", items)
	storage.WriteFile(dir+"/stats.json", stats)
}
