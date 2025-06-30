package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const storage_dir = "storage/"

func WriteFile(name string, content any) {
	var data []byte
	var err error

	switch v := content.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		data = []byte(fmt.Sprint(v))
	default:
		data, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(fmt.Errorf("неизвестный тип и ошибка JSON-маршалинга: %w", err))
		}
	}

	dir := filepath.Dir(storage_dir + name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Errorf("ошибка создания папки: %w", err))
	}

	err = os.WriteFile(storage_dir+name, data, 0644)
	if err != nil {
		panic(fmt.Errorf("ошибка записи файла: %w", err))
	}
}

func ReadFile(filename string) string {
	data, err := os.ReadFile(storage_dir + filename)

	if err != nil {
		panic(err)
	}

	return string(data)
}
