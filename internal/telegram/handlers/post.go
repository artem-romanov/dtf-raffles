package telegram_handlers

import (
	"context"
	telegram_utils "dtf/game_draw/internal/telegram/utils"
	"dtf/game_draw/internal/usecases"
	"log/slog"
	"time"

	"gopkg.in/telebot.v4"
)

type TelegramPostHandlers struct {
	activeRafflesUseCase *usecases.GetActiveRafflePostsUseCase
}

func NewTelegramPostHandlers(
	activeRafflesUseCase *usecases.GetActiveRafflePostsUseCase,
) *TelegramPostHandlers {
	return &TelegramPostHandlers{
		activeRafflesUseCase: activeRafflesUseCase,
	}
}

func (h *TelegramPostHandlers) GetTodayRaffles(ctx telebot.Context) error {
	// telebot doesn't provide context.Context
	// hope in the future it will be availble
	posts, err := h.activeRafflesUseCase.Execute(context.TODO(), time.Now().AddDate(0, 0, -1))
	if err != nil {
		slog.Error("Get active raffles telegram error", "error", err)
		return ctx.Send("Прости друг, не смог достать новости. Попробуй позже.")
	}

	if len(posts) == 0 {
		return ctx.Send("За сегодня не было розыгрышей")
	}

	response := telegram_utils.ManyPostsToTelegramText(posts)
	return ctx.Send(response, telebot.NoPreview)
}
