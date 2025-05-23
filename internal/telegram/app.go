package telegram

import (
	"dtf/game_draw/internal/domain/repositories"
	telegram_handlers "dtf/game_draw/internal/telegram/handlers"
	telegram_middlewares "dtf/game_draw/internal/telegram/middlewares"
	"dtf/game_draw/internal/usecases"
	"fmt"
	"time"

	tele "gopkg.in/telebot.v4"
)

func NewBot(
	botToken string,
	telegramSessionRepo repositories.TelegramSubscribersRepository,
	activeRafflesUseCase *usecases.GetActiveRafflePostsUseCase,
) (*tele.Bot, error) {
	botSettings := tele.Settings{
		ParseMode: tele.ModeHTML,
		Token:     botToken,
		Poller: &tele.LongPoller{
			Timeout: 10 * time.Second,
		},
	}

	bot, err := tele.NewBot(botSettings)
	if err != nil {
		return nil, err
	}

	initalizeMiddlewares(bot)
	err = setCommands(bot)
	if err != nil {
		panic(fmt.Sprintf("Can't set commands. Reason: %s", err.Error()))
	}

	authHandlers := telegram_handlers.NewTelegramAuthHandlers(
		telegramSessionRepo,
	)
	postHandlrs := telegram_handlers.NewTelegramPostHandlers(
		activeRafflesUseCase,
	)

	bot.Handle("/start", func(ctx tele.Context) error {
		return ctx.Send("Попробуй зарегаться, чмо")
	})

	bot.Handle("/subscribe", authHandlers.Subscribe)
	bot.Handle("/unsubscribe", authHandlers.Unsubscribe)

	bot.Handle("/today_raffles", postHandlrs.GetTodayRaffles)

	return bot, nil
}

func initalizeMiddlewares(bot *tele.Bot) *tele.Bot {
	bot.Use(
		telegram_middlewares.RecoverMiddleware,
	)
	return bot
}

func setCommands(bot *tele.Bot) error {
	commands := []tele.Command{
		{
			Text:        "/subscribe",
			Description: "Подписаться на уведомления",
		},
		{
			Text:        "/unsubscribe",
			Description: "Отписаться от уведомлений и уничтожить свою ДТФ сессию",
		},
		{
			Text:        "/today_raffles",
			Description: "Получить список последних розыгрышей",
		},
	}

	err := bot.SetCommands(commands)
	if err != nil {
		return err
	}
	return err
}
