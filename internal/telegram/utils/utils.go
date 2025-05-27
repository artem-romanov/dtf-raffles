package telegram_utils

const (
	MaxPostCharacters = 4096
)

func IsTooLongForTelegramPost(s string) bool {
	strLength := len([]rune(s))
	if strLength > MaxPostCharacters {
		return true
	}
	return false
}
