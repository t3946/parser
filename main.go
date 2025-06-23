package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

type SERPItem struct {
	Pos    int    `json:"pos"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	query := flag.String("query", "", "Search query (e.g., 'купить машину')")
	lr := flag.String("lr", "213", "Region code (e.g., 213 for Moscow)")
	flag.Parse()

	msid := generateMSID()

	if *query == "" {
		log.Fatal("query is required")
	}

	log.Printf("Starting Yandex SERP fetch for query='%s' in region='%s'...", *query, *lr)

	ctx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		[]chromedp.ExecAllocatorOption{
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.UserAgent(randomUserAgent()),
			chromedp.Flag("accept-lang", "ru-RU,ru;q=0.9,en;q=0.8"),
			chromedp.Flag("window-size", "1920,1080"),
			chromedp.Flag("start-maximized", true),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
			chromedp.Flag("enable-automation", false),
		}...,
	)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	results := make([]SERPItem, 0)

	searchURL := fmt.Sprintf("https://yandex.ru/search/?text=%s&lr=%s&primary_reqid=%s", *query, *lr, msid)
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(time.Second),
		chromedp.Evaluate(`window.scrollBy(0, document.body.scrollHeight)`, nil),
		chromedp.WaitVisible("#search-result", chromedp.ByID),
	)
	if err != nil {
		saveErrorPage(ctx, 1)
		log.Fatalf("Failed to load page 1: %v", err)
	}

	for page := 1; page <= 10; page++ {
		if page > 1 {
			script := fmt.Sprintf(`
				(function(){
					const link = document.querySelector("a.Pager-Item[href*='&p=%d']");
					if (link) {
						link.click();
						return true;
					}
					return false;
				})()
			`, page-1)

			var linkClicked bool
			err := chromedp.Run(ctx,
				chromedp.Evaluate(script, &linkClicked),
				chromedp.Sleep(2*time.Second),
			)
			if err != nil || !linkClicked {
				log.Printf("Failed to navigate to page %d (link missing or click failed)", page)
				saveErrorPage(ctx, page)
				continue
			}
		}

		var html string
		err := chromedp.Run(ctx,
			chromedp.WaitReady("#search-result", chromedp.ByID),
			chromedp.OuterHTML("html", &html),
		)
		if err != nil {
			log.Printf("Failed to fetch page %d: %v", page, err)
			saveErrorPage(ctx, page)
			continue
		}

		parsed := extractFromJS(ctx, (page-1)*10)
		results = append(results, parsed...)
		time.Sleep(time.Duration(rand.Intn(3000)+2000) * time.Millisecond)
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("results.json", data, 0644)
	log.Printf("Saved %d results to results.json", len(results))
}

func saveErrorPage(ctx context.Context, page int) {
	tempCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	tempCtx, timeout := context.WithTimeout(tempCtx, 5*time.Second)
	defer timeout()

	var failedHTML string
	var failedShot []byte
	err := chromedp.Run(tempCtx,
		chromedp.OuterHTML("html", &failedHTML),
		chromedp.CaptureScreenshot(&failedShot),
	)
	if err != nil {
		log.Printf("Failed to save error page %d: %v", page, err)
		return
	}
	if failedHTML != "" {
		_ = os.WriteFile(fmt.Sprintf("error_page_%d.html", page), []byte(failedHTML), 0644)
	}
	if len(failedShot) > 0 {
		_ = os.WriteFile(fmt.Sprintf("error_page_%d.png", page), failedShot, 0644)
	}
	if failedHTML == "" && len(failedShot) == 0 {
		log.Printf("Warning: error_page_%d.* files are empty (session likely dead)", page)
	}
}

func generateMSID() string {
	timestamp := time.Now().UnixNano()
	randPart := rand.Uint64()
	dcs := []string{"klg", "vla", "msk", "sas", "iva", "ekb", "spb"}
	dc := dcs[rand.Intn(len(dcs))]
	n := rand.Intn(100)
	return fmt.Sprintf("%d-%d-balancer-l7leveler-kubr-yp-%s-%d-BAL", timestamp, randPart, dc, n)
}

func extractFromJS(ctx context.Context, offset int) []SERPItem {
	var raw []map[string]interface{}
	err := chromedp.Run(ctx, chromedp.Evaluate(`Array.from(document.querySelectorAll('li.serp-item:not(:has(.AdvLabel-Text))')).map((el, i) => {
	const a = el.querySelector('a.Link');
	const titleEl = el.querySelector('.OrganicTitleContentSpan');
	const textEl = el.querySelector('.OrganicTextContentSpan');
	return {
		pos: i + 1,
		url: a?.href || '',
		domain: a?.hostname || '',
		title: titleEl?.innerText || '',
		text: textEl?.innerText || '',
	};
})`, &raw))
	if err != nil {
		log.Printf("Evaluate error: %v", err)
		return nil
	}
	var results []SERPItem
	for _, r := range raw {
		pos := offset + int(r["pos"].(float64))
		results = append(results, SERPItem{
			Pos:    pos,
			URL:    r["url"].(string),
			Domain: r["domain"].(string),
			Title:  r["title"].(string),
			Text:   r["text"].(string),
		})
	}
	return results
}

func randomUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.3034.24 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.7760.16 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.9211.67 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.9836.83 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6942.2 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.4380.99 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.8119.67 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.1303.37 Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.6737.90 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.3302.84 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.8937.70 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.3399.56 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.5815.82 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6323.98 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.5224.76 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.1918.35 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.9172.40 Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.1447.14 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.4413.95 Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.1381.25 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.2607.10 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.1074.70 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.8175.92 Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.7698.16 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.1199.50 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.2641.81 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.1041.92 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.4233.16 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.4228.96 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.1418.46 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.9889.80 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.9953.86 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.4539.2 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.4320.31 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.4869.16 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6145.93 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.3063.32 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.6425.29 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.7656.23 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.7563.84 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.2424.31 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.7919.13 Mobile Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.1471.96 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.8301.91 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.2929.8 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.5330.32 Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.9166.57 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.9694.73 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.6147.79 Mobile Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.9478.13 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.9197.77 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.7833.9 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.9880.69 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.1361.73 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.8259.3 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.7531.50 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.1880.47 Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.9909.93 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.7292.38 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.6109.45 Mobile Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.4453.89 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.7378.86 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.6263.21 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.4711.56 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.3416.94 Mobile Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.8922.44 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.9337.73 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.4060.70 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.2227.32 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.9446.57 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.3855.77 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.3126.50 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.3584.90 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.7697.55 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.8890.97 Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.9976.96 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.8482.67 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.6750.9 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.9624.41 Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.6667.17 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.7458.53 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.1892.86 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.8133.64 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.4513.20 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.8171.25 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.1745.39 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.5077.97 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.7443.68 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.5795.74 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; SM-A536U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.7146.48 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.2098.99 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_3 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.9066.73 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.5214.24 Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.7421.94 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.7344.12 Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.9397.39 Mobile Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.8808.36 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.4619.56 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.9580.34 Safari/537.36",
	}
	return agents[rand.Intn(len(agents))]
}
