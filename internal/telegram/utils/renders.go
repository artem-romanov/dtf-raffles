package telegram_utils

import (
	"dtf/game_draw/internal/domain/models"
	"fmt"
	"strings"
)

func PostToTelegramText(post models.Post) string {
	header := fmt.Sprintf("<b>Розыгрыш</b>: %s\n", post.Title)
	description := fmt.Sprintf("<b>Описание:</b>\n<blockquote expandable>%s</blockquote>", post.Text)
	link := fmt.Sprintf("<b>Ссылка:</b> %s", post.Uri)

	return header + description + link
}

func ManyPostsToTelegramText(posts []models.Post) string {
	builder := strings.Builder{}
	for i, post := range posts {
		if i > 0 && i < len(posts) {
			builder.WriteString("\n * * * \n")
		}
		text := PostToTelegramText(post)
		builder.WriteString(text)
	}
	return builder.String()
}
