/**
 * package browserCtl
 *
 * Browser Control. Package for downloading html pages.
 */

package browserCtl

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"parser/services/config"
	"parser/services/proxy"
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
	Proxy *proxy.Proxy
}

func GetContext(parent context.Context, options GetContextOptions) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent(useragent.RandomUserAgent()),
		chromedp.Flag("accept-lang", "ru-RU,ru;q=0.9,en;q=0.8"),
		chromedp.Flag("window-size", "960,640"),
		chromedp.Flag("start-maximized", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	if options.Proxy != nil {
		urlStr := fmt.Sprintf("http://%s:%s", options.Proxy.Host, options.Proxy.Port)
		opts = append(opts, chromedp.ProxyServer(urlStr))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)

	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	if options.Proxy != nil {
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			go func() {
				switch ev := ev.(type) {
				case *fetch.EventAuthRequired:
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)

					resp := &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: options.Proxy.User,
						Password: options.Proxy.Pass,
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
