package searchYandex

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"net/url"
	browserCtl "parser/services/browserctl"
	"parser/services/httpRequest"
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

func getHeaders() map[string]string {
	return map[string]string{
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"DNT":                       "1",
		"Referer":                   "https://yandex.ru/",
		"Accept-Encoding":           "gzip, deflate, br, zstd",
	}
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

func SearchPhrase(text string, lr string, session Session) []SERPItem {
	result := []SERPItem{}

	for page := 0; page < browserCtl.MaxPage; page++ {
		url := GetSearchPageUrl(text, lr, 0)
		headers := getHeaders()
		headers["Cookie"] = CookieToString(session.Cookie)
		options := map[string]map[string]string{"headers": headers}
		html, _ := httpRequest.Get(url, options)
		result = append(result, ParsePage(html, page)...)
	}

	return result
}

func ParseKeywordsList(keywords []string, lr string) []SERPItem {
	result := []SERPItem{}
	session := GenerateSession(keywords[0], lr)

	for _, keyword := range keywords {
		parsed := SearchPhrase(keyword, lr, session)
		result = append(result, parsed...)
	}

	return result
}
