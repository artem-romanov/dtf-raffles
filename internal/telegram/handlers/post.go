package telegram_handlers

import (
	"context"
	telegram_utils "dtf/game_draw/internal/telegram/utils"
	"dtf/game_draw/internal/usecases"
	"errors"
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

	response := telegram_utils.ManyPostsToTelegramText(posts, false)
	// there is a possibility that number of chars will be greater than 4096
	// and it will screw up sending with "message is too long (400)" error
	// in that case lets send shorten version without description
	if telegram_utils.IsTooLongForTelegramPost(response) {
		response = telegram_utils.ManyPostsToTelegramText(posts, true)
	}
	if err = ctx.Send(response, telebot.NoPreview); err == nil {
		// happy path ends
		return nil
	}

	// not so happy path, lets check whats wrong
	if errors.Is(err, telebot.ErrTooLongMessage) {
		slog.Error("GetTodayRaffles error: Raffles too long:", "length", len([]rune(response)))
		response = telegram_utils.ManyPostsToTelegramText(posts, true)
		if err := ctx.Send(response); err == nil {
			// ok, everyone happy exiting...
			return nil
		}
	}
	// still something wrong? fuck...
	slog.Error("GetTodayRaffles unknown error:", "error", err)
	return ctx.Send("Ошибка. Что-то пошло не так.")

}
