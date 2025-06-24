package searchYandex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"math/rand"
	"os"
	browserCtl "parser/services/browserctl"
	"strings"
	"time"
)

// \Search Engine Results Page Item
type SERPItem struct {
	Pos    int    `json:"pos"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

const CaptchaError = "Captcha error"

// loadPage загружает страницу по указанному URL с использованием переданного контекста.
//
// Параметры:
//   - ctx: контекст для управления временем жизни операции и отмены.
//   - url: строка с адресом страницы, которую необходимо загрузить.
//
// Возвращает:
//   - string: содержимое загруженной страницы в виде строки.
//   - error: ошибку, если загрузка страницы не удалась.
//
// Пример использования:
//
//	content, err := loadPage(ctx, "https://example.com")
//	if err != nil {
//	    log.Fatalf("Ошибка загрузки страницы: %v", err)
//	}
//	fmt.Println("Содержимое страницы:", content)
func loadPage(ctx context.Context, url string) (string, error) {
	var locationHref string
	var html string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second),
		chromedp.Evaluate(`window.scrollBy(0, document.body.scrollHeight)`, nil),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(
			fmt.Sprintf(`document.location.href`),
			&locationHref,
		),
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return "", err
	}

	if strings.Contains(locationHref, "showcaptcha") {
		return "", errors.New(CaptchaError)
	}

	return html, nil
}

func Search(query string, lr string) {
	log.Printf("[INFO] Starting Yandex SERP fetch for query='%s' in region='%s'...", query, lr)

	ctx, cancelAll := browserCtl.GetContext(context.Background())
	defer cancelAll()

	//load keyword page
	msid := generateMSID()
	searchURL := fmt.Sprintf("https://yandex.ru/search/?text=%s&lr=%s&primary_reqid=%s", query, lr, msid)
	_, err := loadPage(ctx, searchURL)

	//handle error
	if err != nil {
		saveErrorPage(ctx, 1)
		log.Fatalf("[ERROR] Failed to load page 1: %v", err)
		return
	}

	results := make([]SERPItem, 0)

	for page := 1; page <= browserCtl.MaxPage; page++ {
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

		parsed := extractFromJS(ctx, (page-1)*browserCtl.MaxPage)
		results = append(results, parsed...)

		//load next page delay
		time.Sleep(time.Duration(rand.Intn(3000)+2000) * time.Millisecond)
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("results.json", data, 0644)
	log.Printf("Saved %d results to results.json", len(results))
}

func generateMSID() string {
	timestamp := time.Now().UnixNano()
	randPart := rand.Uint64()
	dcs := []string{"klg", "vla", "msk", "sas", "iva", "ekb", "spb"}
	dc := dcs[rand.Intn(len(dcs))]
	n := rand.Intn(100)
	return fmt.Sprintf("%d-%d-balancer-l7leveler-kubr-yp-%s-%d-BAL", timestamp, randPart, dc, n)
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
		log.Printf("[ERROR] Failed to save error page %d: %v", page, err)
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
