package telegram_utils

import "fmt"

const (
	MaxPostCharacters = 4096
)

func IsTooLongForTelegramPost(s string) bool {
	strLength := len([]rune(s))
	fmt.Println("checking", strLength)
	if strLength > MaxPostCharacters {
		return true
	}
	return false
}
