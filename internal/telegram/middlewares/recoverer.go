package telegram_middlewares

import (
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

func RecoverMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		defer func() {
			r := recover()
			if r != nil {
				slog.Error("Recoverd from panic: ", "err", r)
			}
		}()
		return next(c)
	}
}
