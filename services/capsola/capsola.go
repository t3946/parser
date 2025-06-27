package capsola

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type createResponse struct {
	Status   int    `json:"status"`
	Response string `json:"response"`
}

func SmartCaptchaCreateTask(clickImageUrl string, taskImageUrl string) string {
	res, err := createTask(map[string]string{
		"type":  "SmartCaptcha",
		"click": getImageBase64(clickImageUrl),
		"task":  getImageBase64(taskImageUrl),
	})

	if err != nil {
		log.Fatal(err)
	}

	var data createResponse

	json.Unmarshal([]byte(res), &data)

	return data.Response
}

func getImageBase64(pathOrUrl string) string {
	var data []byte
	var err error

	if strings.HasPrefix(pathOrUrl, "http://") || strings.HasPrefix(pathOrUrl, "https://") {
		resp, err := http.Get(pathOrUrl)
		if err != nil {
			log.Fatalf("Ошибка загрузки изображения по URL: %v", err)
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Ошибка чтения данных из ответа: %v", err)
		}
	} else {
		// Читаем из локального файла
		data, err = os.ReadFile(pathOrUrl)
		if err != nil {
			log.Fatalf("Ошибка чтения локального файла: %v", err)
		}
	}

	return base64.StdEncoding.EncodeToString(data)
}
