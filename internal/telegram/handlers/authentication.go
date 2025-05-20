package telegram_handlers

import (
	"context"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/repositories"
	telegram_utils "dtf/game_draw/internal/telegram/utils"
	"errors"
	"log/slog"

	tele "gopkg.in/telebot.v4"
)

type TelegramAuthHandlers struct {
	telegramSessionRepo repositories.TelegramSubscribersRepository
}

func NewTelegramAuthHandlers(
	telegramSessionRepo repositories.TelegramSubscribersRepository,
) TelegramAuthHandlers {
	return TelegramAuthHandlers{
		telegramSessionRepo: telegramSessionRepo,
	}
}

func (h *TelegramAuthHandlers) Subscribe(ctx tele.Context) error {
	user := ctx.Sender()
	if user == nil {
		return ctx.Send(telegram_utils.ErrTextUserNotFound)
	}

	// TODO: check telebot docs, look for context
	err := h.telegramSessionRepo.RegisterUser(context.TODO(), nil, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrTelegramUserExists) {
			return ctx.Send("Ошибка. Вы уже зарегестрированы.")
		}
		slog.Error("telegram subscription creation failed. reason: ", "error", err)
		return ctx.Send(telegram_utils.ErrTextUnknown)
	}

	return ctx.Send("Готов. Теперь я буду присылать тебе обновления.")
}

func (h *TelegramAuthHandlers) Unsubscribe(ctx tele.Context) error {
	user := ctx.Sender()
	if user == nil {
		return ctx.Send(telegram_utils.ErrTextUserNotFound)
	}

	// TODO: check telebot docs, look for context
	err := h.telegramSessionRepo.UnregisterUser(context.TODO(), nil, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrTelegramUserNotFound) {
			return ctx.Send("Ошибка. Ты не был зарегестрирован, нечего удалять.")
		}
		return ctx.Send(telegram_utils.ErrTextUnknown)
	}

	// TODO: Remove dtf session too
	// there should be use case with transaction....

	return ctx.Send("Готово. Больше этот бред получать не будешь.")
}
