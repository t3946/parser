package searchYandex

import (
	"context"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"log"
	browserCtl "parser/services/browserctl"
	"parser/services/capsola"
	"parser/services/geometry"
	"strings"
	"time"
)

type Session struct {
	Cookie []*network.Cookie
}

func GenerateSession(text string, lr string) Session {
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

		// modify relative coords to absolute coords
		for i := 0; i < len(solution_coords); i++ {
			solution_coords[i].X += rect.Left
			solution_coords[i].Y += rect.Top
		}

		chromedp.Run(ctx,
			chromedp.MouseClickXY(solution_coords[0].X, solution_coords[0].Y),
			chromedp.Sleep(time.Second),
			chromedp.MouseClickXY(solution_coords[1].X, solution_coords[1].Y),
			chromedp.Sleep(time.Second),
			chromedp.MouseClickXY(solution_coords[2].X, solution_coords[2].Y),
			chromedp.Sleep(time.Second),
			chromedp.MouseClickXY(solution_coords[3].X, solution_coords[3].Y),
			chromedp.Sleep(time.Second),
			chromedp.Evaluate("document.querySelector('.CaptchaButton-ProgressWrapper').click()", &res),
		)
		//[end]
	}

	// Получение всех cookies
	var cookies []*network.Cookie

	chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = storage.GetCookies().Do(ctx)
		return err
	}))

	return Session{
		Cookie: cookies,
	}
}

func CookieToString(cookie []*network.Cookie) string {
	var cookiePairs []string

	for _, c := range cookie {
		cookiePairs = append(cookiePairs, c.Name+"="+c.Value)
	}

	return strings.Join(cookiePairs, "; ")
}
