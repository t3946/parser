package storage

import "os"

func WriteFile(name string, string string) {
	err := os.WriteFile("storage/"+name, []byte(string), 0644)

	if err != nil {
		panic(err)
	}
}
