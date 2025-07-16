package httpRequest

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"errors"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/andybalholm/brotli"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func Get(pageUrl string, options map[string]map[string]string) (string, *http.Response, error) {
	// send req
	req, _ := http.NewRequest("GET", pageUrl, nil)

	for h, v := range options["headers"] {
		req.Header.Add(h, v)
	}

	var client *http.Client

	if options["proxy"] != nil {
		proxyURL, _ := url.Parse(options["proxy"]["proxyStr"])

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		client = &http.Client{
			Transport: transport,
		}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	// read res
	body, err := readResponseBody(resp)

	if err != nil {
		return "", resp, err
	}

	return body, resp, err
}

func GetCycleTls(pageUrl string, options *map[string]map[string]string) (string, *cycletls.Response, error) {
	// send req
	client := cycletls.Init()

	cycletlsOptions := cycletls.Options{
		Body:      "",
		Ja3:       getRandomJA3(),
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	}

	if options != nil {
		headers, ok := (*options)["headers"]

		if ok {
			cycletlsOptions.Headers = headers
		}

		proxy, ok := (*options)["proxy"]

		if ok {
			cycletlsOptions.Proxy = proxy["proxyStr"]
		}
	}

	resp, err := client.Do(pageUrl, cycletlsOptions, "GET")

	if err != nil {
		return "", &resp, err
	}

	// read res
	body := resp.Body

	return body, &resp, err
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

func getRandomJA3() string {
	var ja3List = []string{
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,11-5-51-65037-23-0-45-65281-27-13-18-35-16-43-10,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,16-65037-5-43-10-65281-35-18-51-0-23-45-11-27-13-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,51-10-43-65281-35-18-5-0-16-65037-13-45-11-23-27,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,23-51-16-5-65281-11-35-45-65037-10-13-43-27-0-18-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,5-45-35-65037-18-11-10-27-51-0-16-23-43-13-65281-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,18-16-27-23-51-45-43-65037-65281-5-10-11-13-35-0-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,35-51-45-13-5-0-11-27-23-65037-18-10-16-65281-43-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,35-45-43-16-23-27-5-51-65281-10-11-13-0-65037-18-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,23-43-35-45-51-11-18-10-13-65037-16-5-0-65281-27-41,4588-29-23-24,0",
		"771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,35-45-10-23-11-5-51-65281-43-0-16-13-18-27-65037-41,4588-29-23-24,0",
	}

	rand.Seed(time.Now().UnixNano())
	return ja3List[rand.Intn(len(ja3List))]
}
