package capsola

import "context"

const (
    smartCaptcha = "smartCaptcha"
)

func Solve(ctx context.Context) {
    captchaType := smartCaptcha

    switch captchaType {
    case smartCaptcha:
        //todo: solve captcha
        break
    }
}
