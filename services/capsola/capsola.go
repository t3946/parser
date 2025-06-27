package capsola

import (
    "encoding/base64"
    "io"
    "log"
    "net/http"
    "os"
    "parser/services/httpRequest"
    "strings"
)

func SolveSmartCaptcha(clickImageUrl string, taskImageUrl string) {
    res, err := post(map[string]string{
        "type":  "SmartCaptcha",
        "click": getImageBase64(clickImageUrl),
        "task":  getImageBase64(taskImageUrl),
    })

    if err != nil {
        log.Fatal(err)
    }

    log.Printf(res)
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

func post(data map[string]string) (string, error) {
    url := "https://httpbin.org/post"
    options := map[string]map[string]string{
        "data": data,
        "headers": {
            "Content-Type": "application/json",
            "X-API-Key":    os.Getenv("CAPSOLA_API_KEY"),
        },
    }

    return httpRequest.Post(url, options)
}
