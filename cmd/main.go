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

	"parser/services/useragent"
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
			chromedp.UserAgent(useragent.RandomUserAgent()),
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
