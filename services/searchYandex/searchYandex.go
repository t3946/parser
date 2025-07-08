package searchYandex

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"log"
	"math/rand"
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

type Stats struct {
	TotalPages         int    `json:"total_pages_loaded"`
	TotalCaptchaSolved int    `json:"total_captcha_solved"`
	TimeSpend          string `json:"time_spent"`
}

const CaptchaError string = "Captcha error"

func generateMSID() string {
	timestamp := time.Now().UnixNano()
	randPart := rand.Uint64()
	dcs := []string{"klg", "vla", "msk", "sas", "iva", "ekb", "spb"}
	dc := dcs[rand.Intn(len(dcs))]
	n := rand.Intn(100)
	return fmt.Sprintf("%d-%d-balancer-l7leveler-kubr-yp-%s-%d-BAL", timestamp, randPart, dc, n)
}

// GetSearchPageUrl формирует URL для поиска на Яндексе с заданным текстом, регионом и номером страницы.
// Если номер страницы больше 0, добавляется параметр пагинации "p".
// Возвращает готовую строку URL с корректно закодированными
func GetSearchPageUrl(text string, lr string, page int) string {
	pageUrl, _ := url.Parse("https://yandex.ru/search/")
	params := url.Values{}
	params.Add("text", text)
	params.Add("lr", lr)
	params.Add("msid", generateMSID())

	if page > 0 {
		params.Add("p", strconv.Itoa(page))
	}

	pageUrl.RawQuery = params.Encode()

	return pageUrl.String()
}

// LoadPage загружает страницу по указанному URL с использованием переданного контекста.
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
// content, err := LoadPage(ctx, "https://example.com")
//
//	if err != nil {
//	    log.Fatalf("Ошибка загрузки страницы: %v", err)
//	}
//
// fmt.Println("Содержимое страницы:", content)
func LoadPage(ctx context.Context, url string, session *Session) (string, error) {
	var locationHref string
	var html string
	var err error

	if session != nil {
		err = chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				return browserCtl.SetCookiesFromNetworkCookies(ctx, session.Cookie)
			}),
		)
	}

	err = chromedp.Run(ctx,
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

func sleep(min, max float64) {
	if min > max {
		log.Fatalf("invalid arguments: min (%f) > max (%f)", min, max)
	}

	// Создаём локальный генератор с уникальным сидом
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	minInt := int(min * 10)
	maxInt := int(max * 10)

	durationTenths := r.Intn(maxInt-minInt+1) + minInt
	duration := float64(durationTenths) / 10.0

	log.Printf("Sleeping for %.1f seconds...", duration)
	time.Sleep(time.Duration(duration * float64(time.Second)))
}

func GetHeaders() map[string]string {
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

func ParseKeywordsList(keywords []string, lr string) ([]SERPItem, Stats) {
	result := []SERPItem{}
	solvedCaptchaTotal := 0
	session, solvedCaptcha := GenerateSession(keywords[0], lr, nil)
	solvedCaptchaTotal += solvedCaptcha
	startTime := time.Now()
	totalPages := 0

	for _, keyword := range keywords {
		log.Printf("[INFO] Parse KW: `%v`", keyword)
		parsed := []SERPItem{}

		for page := 0; page < browserCtl.MaxPage; page++ {
			url := GetSearchPageUrl(keyword, lr, page)
			log.Printf("[INFO]     Url: `%v`", url)
			headers := GetHeaders()
			headers["Cookie"] = CookieToString(session.Cookie)
			options := map[string]map[string]string{
				"headers": headers,
			}

			if browserCtl.UseProxy {
				options["proxy"] = map[string]string{
					"proxyStr": "http://C3smQv:FPQoP8bkSX@77.83.148.95:1050",
				}
			}

			html, resp, _ := httpRequest.Get(url, options)

			if strings.Contains(resp.Request.URL.String(), "showcaptcha") {
				page -= 1
				session, solvedCaptcha = GenerateSession(keyword, lr, &session)
				solvedCaptchaTotal += solvedCaptcha
				continue
			}

			log.Printf("[INFO]     Append KW to parsed")
			parsed = append(parsed, ParsePage(html, page)...)
			totalPages += 1
			sleep(4, 6)
		}

		result = append(result, parsed...)
	}

	elapsed := time.Since(startTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	stats := Stats{
		TotalPages:         totalPages,
		TotalCaptchaSolved: solvedCaptchaTotal,
		TimeSpend:          fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds),
	}

	return result, stats
}

func ParseKeywordsListWithChromeDP(keywords []string, lr string) ([]SERPItem, Stats) {
	result := []SERPItem{}
	solvedCaptchaTotal := 0
	ctx, cancelAll := browserCtl.GetContext(context.Background())
	defer cancelAll()
	startTime := time.Now()
	totalPages := 0
	var session Session

	for _, keyword := range keywords {
		parsed := []SERPItem{}

		for page := 0; page < browserCtl.MaxPage; page++ {
			log.Printf("[INFO] Parse KW: `%v[%v]`", keyword, page)
			url := GetSearchPageUrl(keyword, lr, page)
			html, err := LoadPage(ctx, url, &session)

			if err == nil {
				log.Printf("[INFO]     Append KW to parsed")
				parsed = append(parsed, ParsePage(html, page)...)
				totalPages += 1
				sleep(4, 6)
			} else if err.Error() == CaptchaError {
				page -= 1
				solvedCaptchaTotal += SolveCaptcha(ctx)
			} else {
				panic(err)
			}
		}

		result = append(result, parsed...)
	}

	elapsed := time.Since(startTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	stats := Stats{
		TotalPages:         totalPages,
		TotalCaptchaSolved: solvedCaptchaTotal,
		TimeSpend:          fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds),
	}

	return result, stats
}
