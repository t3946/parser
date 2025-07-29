package searchYandex

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"log"
	browserCtl "parser/services/browserctl"
	"parser/services/capsola"
	"parser/services/config"
	"parser/services/geometry"
	"parser/services/proxy"
	"strings"
	"time"
)

type Session struct {
	Cookie []*network.Cookie
}

func GenerateSession(text string, lr string, oldSession *Session) (Session, int, error) {
	if oldSession != nil {
		log.Printf("[INFO] Retrust session")
	} else {
		log.Printf("[INFO] Generate new session")
	}

	contextOptions := browserCtl.GetContextOptions{
		Proxy: nil,
	}

	if config.UseProxy {
		proxyStruct := proxy.GetProxy(true)
		contextOptions.Proxy = &proxyStruct
	}

	ctx, cancelAll := browserCtl.GetContext(context.Background(), contextOptions)
	defer cancelAll()

	_, err := LoadPage(ctx, GetSearchPageUrl(text, lr, 0), oldSession)

	var solvedCaptcha = 0

	if err != nil {
		if err.Error() == CaptchaError {
			solvedCaptcha = SolveCaptcha(ctx)
		} else {
			return Session{}, 0, err
		}
	}

	session := Session{
		Cookie: getCookieFromCtx(ctx),
	}

	return session, solvedCaptcha, nil
}

func SolveCaptcha(ctx context.Context) int {
	var isCaptchaSolved = false
	var solvedCaptchaCount = 0

	for !isCaptchaSolved {
		log.Printf("[INFO] Captcha offered")

		var res interface{}
		var clickImageUrl string
		var taskImageUrl string
		var currentURL string

		chromedp.Run(ctx,
			// click button "i am not a robot"
			chromedp.Evaluate("document.getElementById('js-button')?.click()", &res),
			chromedp.Sleep(time.Second*2),
			chromedp.Location(&currentURL),
		)

		//captcha passed in one click
		if !strings.Contains(currentURL, "showcaptcha") {
			log.Printf("[INFO] Captcha checkbox accepted")
			break
		}

		log.Printf("[INFO] Solve smart captcha")

		chromedp.Run(ctx,
			chromedp.WaitVisible(".AdvancedCaptcha-ImageWrapper img"),
			chromedp.WaitVisible(".AdvancedCaptcha-SilhouetteTask img"),
			chromedp.Evaluate("document.querySelector('.AdvancedCaptcha-ImageWrapper img').src", &clickImageUrl),
			chromedp.Evaluate("document.querySelector('.AdvancedCaptcha-SilhouetteTask img').src", &taskImageUrl),
		)

		//[start] get solution Smart Captcha
		task_id := capsola.SmartCaptchaCreateTask(clickImageUrl, taskImageUrl)

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

		// modify relative coords to absolute coords
		for i := 0; i < len(solution_coords); i++ {
			solution_coords[i].X += rect.Left
			solution_coords[i].Y += rect.Top
		}

		chromedp.Run(ctx,
			chromedp.MouseClickXY(solution_coords[0].X, solution_coords[0].Y),
			chromedp.MouseClickXY(solution_coords[1].X, solution_coords[1].Y),
			chromedp.MouseClickXY(solution_coords[2].X, solution_coords[2].Y),
			chromedp.MouseClickXY(solution_coords[3].X, solution_coords[3].Y),
			chromedp.Evaluate("document.querySelector('.CaptchaButton-ProgressWrapper').click()", &res),
			chromedp.WaitReady("body"),
			chromedp.Sleep(time.Second),
			chromedp.Location(&currentURL),
		)
		//[end]

		if !strings.Contains(currentURL, "showcaptcha") {
			isCaptchaSolved = true
		}

		solvedCaptchaCount++
	}

	return solvedCaptchaCount
}

func CookieToString(cookie []*network.Cookie) string {
	var cookiePairs []string

	for _, c := range cookie {
		cookiePairs = append(cookiePairs, c.Name+"="+c.Value)
	}

	return strings.Join(cookiePairs, "; ")
}

func getCookieFromCtx(ctx context.Context) []*network.Cookie {
	var cookies []*network.Cookie

	chromedp.Run(
		ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = storage.GetCookies().Do(ctx)
			return err
		}),
	)

	return cookies
}
