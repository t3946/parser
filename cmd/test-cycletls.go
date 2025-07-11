package main

import (
	"parser/services/httpRequest"
	"parser/services/storage"
)

func main() {
	respBody, _, _ := httpRequest.GetCycleTls("https://ya.ru/search/?text=%D1%81%D0%BF%D1%83%D1%84%D0%B8%D0%BD%D0%B3&clid=12124976-2&lr=46", nil)

	storage.WriteFile("result.html", respBody)
}
