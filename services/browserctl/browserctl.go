/**
 * package browserCtl
 *
 * Browser Control. Package for downloading html pages.
 */

package browserCtl

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"parser/services/useragent"
)

// Search Engine Results Page Item
type SERPItem struct {
	Pos    int    `json:"pos"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

func GetContext(parent context.Context) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent(useragent.RandomUserAgent()),
		chromedp.Flag("accept-lang", "ru-RU,ru;q=0.9,en;q=0.8"),
		chromedp.Flag("window-size", "960,640"),
		chromedp.Flag("start-maximized", false),
		chromedp.Flag("enable-automation", false),
	)

	if UseProxy {
		opts = append(opts, chromedp.ProxyServer("http://77.83.148.95:1050"))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	UseProxy := false

	if UseProxy {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			go func() {
				switch ev := ev.(type) {
				case *fetch.EventAuthRequired:
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)

					resp := &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: "C3smQv",
						Password: "FPQoP8bkSX",
					}

					err := fetch.ContinueWithAuth(ev.RequestID, resp).Do(execCtx)
					if err != nil {
						log.Print(err)
					}

				case *fetch.EventRequestPaused:
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)
					err := fetch.ContinueRequest(ev.RequestID).Do(execCtx)
					if err != nil {
						log.Print(err)
					}
				}
			}()
		})

		err := chromedp.Run(ctx,
			fetch.Enable().WithHandleAuthRequests(true),
		)

		if err != nil {
			log.Fatal(err)
		}
	}

	cancel := func() {
		cancelCtx()
		cancelAlloc()
	}

	var cancelTimeout context.CancelFunc

	if TimeOutSec > 0 {
		ctx, cancelTimeout = context.WithTimeout(ctx, TimeOutSec)
	}

	cancelAll := func() {
		if cancelTimeout != nil {
			cancelTimeout()
		}

		cancel()
	}

	return ctx, cancelAll
}

func SetCookiesFromNetworkCookies(ctx context.Context, cookies []*network.Cookie) error {
	var cookieParams []*network.CookieParam

	for _, c := range cookies {
		cp := &network.CookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: c.SameSite,
		}

		cookieParams = append(cookieParams, cp)
	}

	return network.SetCookies(cookieParams).Do(ctx)
}
