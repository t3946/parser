package httpRequest

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func Get(url string, options map[string]map[string]string) (string, error) {
	// send req
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", options["headers"]["cookie"])
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	// read res
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа: %v", err)
	}

	return string(body), err
}

func Post(url string, options map[string]map[string]string) (string, error) {
	data, dataExists := options["data"]
	headers, headersExists := options["headers"]

	if !dataExists {
		data = map[string]string{}
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Fatalf("Ошибка сериализации JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	if err != nil {
		log.Fatalf("Ошибка создания запроса: %v", err)
	}

	if headersExists {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// send req
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	// read res
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа: %v", err)
	}

	return string(body), err
}
