/**
 * package browserCtl
 *
 * Browser Control. Package for downloading html pages.
 */

package browserCtl

import (
	"context"
	"github.com/chromedp/chromedp"
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

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(parent, opts...)

	ctx, cancelCtx := chromedp.NewContext(allocCtx)

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
