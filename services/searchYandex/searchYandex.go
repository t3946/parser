package searchYandex

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"log"
	"net/url"
	browserCtl "parser/services/browserctl"
	"parser/services/httpRequest"
	"parser/services/storage"
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

func ParsePage(html string, page int) []SERPItem {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	result := []SERPItem{}
	nodes := doc.Find("li.serp-item:not(:has(.AdvLabel-Text)):not([data-fast-name=\"images\"])")
	nodes.Each(func(i int, node *goquery.Selection) {
		aNode := node.Find("a.Link")
		titleEl := node.Find(".OrganicTitleContentSpan")
		textEl := node.Find(".OrganicTextContentSpan")
		linkUrl, _ := aNode.Attr("href")
		u, _ := url.Parse(linkUrl)

		result = append(result, SERPItem{
			Pos:    i + 1 + page*nodes.Length(),
			URL:    u.String(),
			Domain: u.Hostname(),
			Title:  titleEl.Text(),
			Text:   textEl.Text(),
		})
	})

	return result
}

func SearchPhrase(text string, lr string) {
	log.Printf("[INFO] Starting Yandex SERP fetch for text='%s' in region='%s'...", text, lr)

	session := GenerateSession(text, lr)
	cookie_str := CookieToString(session.Cookie)

	options := map[string]map[string]string{"header": {"cookie": cookie_str}}
	url := GetSearchPageUrl(text, lr, 0)
	html, err := httpRequest.Get(url, options)

	if err != nil {
		log.Println("html not loaded")
		log.Println(err)
	} else {
		log.Println(options)
	}

	storage.WriteFile("page.html", html)

	time.Sleep(time.Second * 3000)

	for i := 1; i < browserCtl.MaxPage; i++ {

	}

	return
}
