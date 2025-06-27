package capsola

import (
	"net/url"
	"os"
	"parser/services/httpRequest"
	"strings"
)

func apiUrl(path string) string {
	base, _ := url.Parse("https://api.capsola.cloud")

	// Убираем ведущие слэши, чтобы корректно добавить путь
	path = strings.TrimLeft(path, "/")

	// Обновляем путь базового URL
	base.Path = base.Path + "/" + path

	return base.String()
}

func getHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    os.Getenv("CAPSOLA_API_KEY"),
	}
}

func createTask(data map[string]string) (string, error) {
	options := map[string]map[string]string{
		"data":    data,
		"headers": getHeaders(),
	}

	return httpRequest.Post(apiUrl("create"), options)
}

func getResult(data map[string]string) (string, error) {
	options := map[string]map[string]string{
		"data":    data,
		"headers": getHeaders(),
	}

	return httpRequest.Post(apiUrl("result"), options)
}
