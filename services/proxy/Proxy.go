package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"parser/services/httpRequest"
	"strings"
	"time"
)

type Proxy struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Host string `json:"host"`
	Port string `json:"port"`
}

var index = 0
var proxies = []Proxy{}
var testURL = "https://yandex.ru"
var testTimeOutSec = 1

func Init() {
	options := map[string]map[string]string{}
	response, _, err := httpRequest.Get("https://api.proxytraff.com/package/get?c=K5hk", options)

	if err != nil {
		panic("Can't load proxy list")
	}

	proxiesStr := strings.Split(response, "\n")

	for _, proxyStr := range proxiesStr {
		proxies = append(proxies, proxyStrToStruct(proxyStr))
	}
}

func structToStr(proxy Proxy) string {
	if proxy.User != "" {
		return fmt.Sprintf("http://%s:%s@%s:%s", proxy.User, proxy.Pass, proxy.Host, proxy.Port)
	}

	return fmt.Sprintf("http://%s:%s", proxy.Host, proxy.Port)
}

func proxyStrToStruct(proxy string) Proxy {
	userinfoAndHost := strings.SplitN(proxy, "@", 2)
	var user = ""
	var pass = ""
	var host = ""
	var port = ""

	if len(userinfoAndHost) == 2 {
		userPass := strings.SplitN(userinfoAndHost[0], ":", 2)
		user = userPass[0]
		pass = userPass[1]

		hostPort := strings.SplitN(userinfoAndHost[1], ":", 2)
		host = hostPort[0]
		port = hostPort[1]
	} else {
		hostPort := strings.SplitN(userinfoAndHost[0], ":", 2)
		host = hostPort[0]
		port = hostPort[1]
	}

	return Proxy{
		User: user,
		Pass: pass,
		Host: host,
		Port: port,
	}
}

func checkProxy(proxy Proxy) bool {
	proxyStr := fmt.Sprintf("%s:%s", proxy.Host, proxy.Port)
	proxyURL, _ := url.Parse(proxyStr)

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(testTimeOutSec) * time.Second,
	}

	resp, err := client.Get(testURL)

	defer resp.Body.Close()

	return err == nil && resp.StatusCode == 200
}

func GetProxy() Proxy {
	var proxy Proxy
	checked := false

	for checked == false {
		proxy = proxies[index]
		checked = checkProxy(proxy)
		index += 1
	}

	return proxy
}
