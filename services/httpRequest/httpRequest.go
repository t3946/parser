package httpRequest

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"github.com/andybalholm/brotli"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func readResponseBody(resp *http.Response) (string, error) {
	var reader io.Reader = resp.Body
	defer resp.Body.Close()

	encoding := strings.ToLower(resp.Header.Get("Content-Encoding"))

	switch encoding {
	case "gzip":
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gzReader.Close()
		reader = gzReader
	case "deflate":
		zReader, err := zlib.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer zReader.Close()
		reader = zReader
	case "br":
		reader = brotli.NewReader(resp.Body)
	case "", "identity":
		// без сжатия
	default:
		return "", errors.New("неподдерживаемый Content-Encoding: " + encoding)
	}

	bodyBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func Get(url string, options map[string]map[string]string) (string, error) {
	// send req
	req, _ := http.NewRequest("GET", url, nil)

	for h, v := range options["headers"] {
		req.Header.Add(h, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	// read res
	body, err := readResponseBody(resp)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа: %v", err)
	}

	return body, err
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
