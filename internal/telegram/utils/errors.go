package telegram_utils

const (
	ErrTextUnknown      = "Неизвестная ошибка. Прости, друг."
	ErrTextUserNotFound = "Ошибка. Юзер (ты) не найден"
)

// sends
// func ValidateUser(ctx tele.Context, user tele.User) error {
// 	if user != nil {
// 		return
// 	}
// 	return ctx.Send("Ошибка. Юзер (ты) не найден.")
// }
