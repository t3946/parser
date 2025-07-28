package main

import (
	"fmt"
	"net/http"
	"net/url"
	"parser/services/httpRequest"
	"strings"
	"sync"
	"time"
)

type TReport struct {
	proxy  string
	status int
}

func checkProxy(proxyAddr, testURL string, timeoutSec int, ch chan TReport) {
	proxyURL, _ := url.Parse(proxyAddr)

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeoutSec) * time.Second,
	}

	resp, err := client.Get(testURL)

	if err != nil {
		ch <- TReport{proxyAddr, -1}
		return
	}

	defer resp.Body.Close()
	ch <- TReport{proxyAddr, resp.StatusCode}
}

func readProxies() []string {
	options := map[string]map[string]string{}
	response, _, _ := httpRequest.Get("https://api.proxytraff.com/package/get?c=K5hk", options)

	return strings.Split(response, "\n")
}

func main() {
	proxies := readProxies()[:10]
	targetURL := "https://yandex.ru"
	timeout := 1
	reports := make(chan TReport)
	sem := make(chan struct{}, 1) // семафор с десятью слотами

	var wg sync.WaitGroup

	for _, proxy := range proxies {
		wg.Add(1)
		go func(proxy string) {
			defer wg.Done()
			sem <- struct{}{} // занять слот (если занято 10 – блокируется)
			checkProxy(proxy, targetURL, timeout, reports)
			<-sem // освободить слот
		}(proxy)
	}

	// горутина для закрытия канала с результатами
	go func() {
		wg.Wait()
		close(reports)
	}()

	for report := range reports {
		fmt.Printf("Прокси %s: %d\n", report.proxy, report.status)
	}
}
