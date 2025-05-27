package telegram_utils

import (
	"dtf/game_draw/internal/domain/models"
	"fmt"
	"strings"
)

func PostToTelegramText(post models.Post, short bool) string {
	sb := strings.Builder{}

	// header
	sb.WriteString(fmt.Sprintf("<b>Розыгрыш</b>: %s\n", post.Title))

	if post.Text != "" && !short {
		sb.WriteString(
			fmt.Sprintf("<b>Описание:</b>\n<blockquote expandable>%s</blockquote>", post.Text),
		)
	}

	sb.WriteString(fmt.Sprintf("<b>Ссылка:</b> %s", post.Uri))

	return sb.String()
}

func ManyPostsToTelegramText(posts []models.Post, short bool) string {
	builder := strings.Builder{}
	for i, post := range posts {
		if i > 0 && i < len(posts) {
			builder.WriteString("\n * * * \n")
		}
		text := PostToTelegramText(post, short)
		builder.WriteString(text)
	}
	if short {
		builder.WriteString("<i>\n\nОписания слишком длинные, укорочено</i>")
	}
	return builder.String()
}
