package capsola

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"parser/services/geometry"
	"strconv"
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

func SmartCaptchaGetSolution(task_id string) []geometry.Point {
	res, err := getResult(map[string]string{
		"id": task_id,
	})

	if err != nil {
		log.Fatal(err)
	}

	var data createResponse

	json.Unmarshal([]byte(res), &data)

	// [start] parse coords str
	coordsStr := strings.Split(data.Response, ":")[1]
	coordsStrPairs := strings.Split(coordsStr, ";")
	coordsList := []geometry.Point{}

	for i := 0; i < len(coordsStrPairs); i++ {
		pair := strings.Split(coordsStrPairs[i], ",")

		X, _ := strconv.ParseFloat(strings.Split(pair[0], "=")[1], 64)
		Y, _ := strconv.ParseFloat(strings.Split(pair[1], "=")[1], 64)

		point := geometry.Point{
			X: X,
			Y: Y,
		}

		coordsList = append(coordsList, point)
	}
	// [end]

	return coordsList
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
