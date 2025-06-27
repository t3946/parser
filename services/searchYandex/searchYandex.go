package searchYandex

import (
	"context"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"math/rand"
	"net/url"
	"os"
	browserCtl "parser/services/browserctl"
	"parser/services/capsola"
	"parser/services/geometry"
	"strconv"
	"strings"
	"time"
)

// \SearchPhrase Engine Results Page Item
type SERPItem struct {
	Pos    int    `json:"pos"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

const CaptchaError string = "Captcha error"

func getStartPageUrl() string {
	return "https://yandex.ru/"
}

// GetSearchPageUrl формирует URL для поиска на Яндексе с заданным текстом, регионом и номером страницы.
// Если номер страницы больше 0, добавляется параметр пагинации "p".
// Возвращает готовую строку URL с корректно закодированными
func GetSearchPageUrl(text string, lr string, page int) string {
	pageUrl, _ := url.Parse("https://yandex.ru/search/")
	params := url.Values{}
	params.Add("text", text)
	params.Add("lr", lr)

	if page > 0 {
		params.Add("p", strconv.Itoa(page))
	}

	pageUrl.RawQuery = params.Encode()

	return pageUrl.String()
}

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
// content, err := loadPage(ctx, "https://example.com")
//
//	if err != nil {
//	    log.Fatalf("Ошибка загрузки страницы: %v", err)
//	}
//
// fmt.Println("Содержимое страницы:", content)
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
		chromedp.Sleep(time.Second*1),
	)

	if err != nil {
		return "", err
	}

	if strings.Contains(locationHref, "showcaptcha") {
		return "", errors.New(CaptchaError)
	}

	return html, nil
}

func parsePage(html string) {
	//раньше делалась эта хрень
	//extractFromJS(ctx, (page-1)*browserCtl.MaxPage)
	//todo: теперь надо заменить на краулер или иной модуль
}

var rect struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func SearchPhrase(text string, lr string) {
	log.Printf("[INFO] Starting Yandex SERP fetch for text='%s' in region='%s'...", text, lr)

	//todo: надо посетить начальную страницу в начале сессии

	ctx, cancelAll := browserCtl.GetContext(context.Background())
	defer cancelAll()

	_, err := loadPage(ctx, GetSearchPageUrl(text, lr, 0))

	if err.Error() == CaptchaError {
		var res interface{}
		var clickImageUrl string
		var taskImageUrl string

		chromedp.Run(ctx,
			// click button "i am not a robot"
			chromedp.Evaluate("document.getElementById('js-button').click()", &res),
			chromedp.Sleep(time.Second*5),
			// todo: if captcha will no appear, this will throw error
			chromedp.Evaluate("document.querySelector('.AdvancedCaptcha-ImageWrapper img').src", &clickImageUrl),
			chromedp.Evaluate("document.querySelector('.AdvancedCaptcha-SilhouetteTask img').src", &taskImageUrl),
		)

		//[start] get solution Smart Captcha
		task_id := capsola.SmartCaptchaCreateTask(clickImageUrl, taskImageUrl)

		log.Printf(task_id)

		time.Sleep(time.Second * 1)

		solution_coords := capsola.SmartCaptchaGetSolution(task_id)
		//[end]

		//[start] solve Smart Captcha
		var rect geometry.Rectangle

		chromedp.Run(ctx,
			chromedp.EvaluateAsDevTools(`
					(function(){
						var el = document.querySelector('.AdvancedCaptcha-ImageWrapper img');
						var r = el.getBoundingClientRect();
						return {left: r.left, top: r.top, width: r.width, height: r.height};
					})()
				`, &rect),
		)

		for i := 0; i < len(solution_coords); i++ {
			//todo: input solution

			//submit solution
			//chromedp.Run(ctx,
			//	chromedp.Evaluate("document.querySelector('.CaptchaButton-ProgressWrapper').click()", &res),
			//)
		}
		//[end]

	}

	time.Sleep(time.Second * 3000)
	for i := 1; i < browserCtl.MaxPage; i++ {

	}

	return
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
