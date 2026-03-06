package telegram_handlers

import (
	"context"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/repositories"
	telegram_utils "dtf/game_draw/internal/telegram/utils"
	"errors"
	"fmt"
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TelegramAuthHandlers struct {
	telegramSessionRepo repositories.TelegramSubscribersRepository
	telegramAdmins      []int64
}

func NewTelegramAuthHandlers(
	telegramSessionRepo repositories.TelegramSubscribersRepository,
	telegramAdmins []int64,
) *TelegramAuthHandlers {
	return &TelegramAuthHandlers{
		telegramSessionRepo: telegramSessionRepo,
		telegramAdmins:      telegramAdmins,
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
			return ctx.Send("⚠️ Ошибка. Ты уже подписан на обновления.")
		}
		slog.Error("telegram subscription creation failed. reason: ", "error", err)
		return ctx.Send(telegram_utils.ErrTextUnknown)
	}

	if err := ctx.Send("✅ Готово!\n\n🔔Теперь я буду присылать тебе обновления каждый день в 14:00"); err != nil {
		return err
	}

	if err := telegram_utils.BroadcastWithRetries(
		context.TODO(),
		ctx.Bot(),
		fmt.Sprintf("🆕 Новый подписчик!\n\nTelegram ID: %v", user.ID),
		h.telegramAdmins,
	); err != nil {
		slog.Error(
			"couldnt notify about new telegram sub",
			"err",
			err,
			"telegram_id",
			user.ID,
		)
	}

	// sleep to avoid ban from telegram
	// TODO: think about it, maybe we should remove it
	time.Sleep(50 * time.Millisecond)

	return nil
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
			return ctx.Send("⚠️ Ошибка. Ты не был подписан, поэтому удалять нечего.")
		}
		return ctx.Send(telegram_utils.ErrTextUnknown)
	}

	// TODO: Remove dtf session too
	// there should be use case with transaction....

	if err := ctx.Send("✅ Готово!\n\n🔕Ты больше не получишь обновления."); err != nil {
		return err
	}

	// sleep to avoid ban from telegram
	// TODO: think about it, maybe we should remove it
	time.Sleep(50 * time.Millisecond)

	if err := telegram_utils.BroadcastWithRetries(
		context.TODO(),
		ctx.Bot(),
		fmt.Sprintf("❌ Пользователь отписался\n\nTelegram ID: %v", user.ID),
		h.telegramAdmins,
	); err != nil {
		slog.Error(
			"couldnt notify about new telegram un-sub",
			"err",
			err,
			"telegram_id",
			user.ID,
		)
	}

	return nil
}
