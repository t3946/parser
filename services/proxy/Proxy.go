package proxy

import (
	"encoding/json"
	"parser/services/storage"
	"strings"
)

type Proxy struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Host string `json:"host"`
	Port string `json:"port"`
}

func readFromFile() string {
	var proxyList []string

	jsonData := storage.ReadFile("proxy.json")
	json.Unmarshal([]byte(jsonData), &proxyList)

	return proxyList[0]
}

func ProxyStrToStruct(proxy string) Proxy {
	userinfoAndHost := strings.SplitN(proxy, "@", 2)

	userPass := strings.SplitN(userinfoAndHost[0], ":", 2)
	user := userPass[0]
	pass := userPass[1]

	hostPort := strings.SplitN(userinfoAndHost[1], ":", 2)
	host := hostPort[0]
	port := hostPort[1]

	return Proxy{
		User: user,
		Pass: pass,
		Host: host,
		Port: port,
	}
}

func GetProxyStr() string {
	return readFromFile()
}

func GetProxy() Proxy {
	proxyStr := GetProxyStr()

	return ProxyStrToStruct(proxyStr)
}
