package telegram_utils

import "unicode/utf8"

const (
	MaxPostCharacters = 4096
)

func IsTooLongForTelegramPost(s string) bool {
	strLength := utf8.RuneCountInString(s)
	return strLength > MaxPostCharacters
}
