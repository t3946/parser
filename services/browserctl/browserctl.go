/**
 * package browserCtl
 *
 * Browser Control. Package for downloading html pages.
 */

package browserCtl

import (
	"context"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"parser/services/config"
	"parser/services/proxyx"
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

type GetContextOptions struct {
	Proxy *proxyx.TProxy
}

func GetContext(parent context.Context, options GetContextOptions) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", config.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent(useragent.RandomUserAgent()),
		chromedp.Flag("accept-lang", "ru-RU,ru;q=0.9,en;q=0.8"),
		chromedp.Flag("window-size", "960,640"),
		chromedp.Flag("start-maximized", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	if config.UseProxy && options.Proxy != nil {
		opts = append(opts,
			chromedp.ProxyServer(proxyx.StructToStr(*options.Proxy)),
			chromedp.Flag("proxy-bypass-list", "<-loopback>"),
		)
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)

	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	if config.UseProxy && options.Proxy != nil && options.Proxy.User != "" {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			switch ev := ev.(type) {
			case *fetch.EventAuthRequired:
				if ev.AuthChallenge.Source == fetch.AuthChallengeSourceProxy {
					go func() {
						err := chromedp.Run(ctx, fetch.ContinueWithAuth(ev.RequestID, &fetch.AuthChallengeResponse{
							Response: fetch.AuthChallengeResponseResponseProvideCredentials,
							Username: options.Proxy.User,
							Password: options.Proxy.Pass,
						}))
						if err != nil {
							log.Printf("auth error: %v", err)
						}
					}()
				}
			case *fetch.EventRequestPaused:
				go func() {
					_ = chromedp.Run(ctx, fetch.ContinueRequest(ev.RequestID))
				}()
			}
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

	if config.TimeOutSec > 0 {
		ctx, cancelTimeout = context.WithTimeout(ctx, config.TimeOutSec)
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
