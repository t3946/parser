package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"math"
	"parser/services/config"
	"parser/services/proxy"
	"parser/services/searchYandex"
	"parser/services/storage"
	"slices"
	"strings"
	"sync"
	"time"
)

func init() {
	godotenv.Load()
	proxy.Init()
}

func main() {
	//fetch sources data
	dataStr := storage.ReadFile("test/10000.txt")
	kw := strings.Split(dataStr, "\n")[0:config.KwNumber]

	//chunk source data
	chunkSize := math.Ceil(float64(len(kw)) / float64(config.Threads))
	chunks := slices.Chunk(kw, int(chunkSize))

	//[start] process input data
	log.Printf("[INFO] Parse %v keyword(s)", config.KwNumber)

	startTime := time.Now()
	resultsCh := make(chan searchYandex.TResult)
	sem := make(chan struct{}, config.Threads) // семафор с десятью слотами

	var wg sync.WaitGroup

	for chunk := range chunks {
		time.Sleep(time.Second * 3)
		wg.Add(1)
		go func(chunk []string) {
			defer wg.Done()
			searchYandex.ParseKeywordsListRoutine(chunk, "46", resultsCh)
			sem <- struct{}{} // block slot

			<-sem // free slot
		}(chunk)
	}

	// close channel
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var items = []searchYandex.SERPItem{}
	var stats = searchYandex.Stats{}

	for result := range resultsCh {
		items = append(items, result.Items...)

		stats.TotalPages += result.Stats.TotalPages
		stats.TotalCaptchaSolved += result.Stats.TotalCaptchaSolved
		stats.AccessSuspended += result.Stats.AccessSuspended
	}

	elapsed := time.Since(startTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	stats.TimeSpend = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	//[end]

	//output results
	dir := fmt.Sprintf("load-kw-test-%v", config.KwNumber)
	storage.WriteFile(dir+"/result.json", items)
	storage.WriteFile(dir+"/stats.json", stats)
}
