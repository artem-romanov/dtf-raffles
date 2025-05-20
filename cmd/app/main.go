package main

import (
	"context"
	"database/sql"
	"dtf/game_draw/internal/domain/models"
	iRepo "dtf/game_draw/internal/domain/repositories"
	"dtf/game_draw/internal/repositories"
	"dtf/game_draw/internal/storage/sqlite"
	"dtf/game_draw/internal/telegram"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/telebot.v4"
)

const (
	sqlitePath = "./dtf_db.sqlite"
)

func main() {
	deps := initDependencies(sqlitePath)

	bot, err := telegram.NewBot(deps.telegramSubsRepo)
	if err != nil {
		log.Fatalf("Fuck! Reason: %s", err)
	}

	schedulder := setupScheduledJobs(bot, deps.telegramSubsRepo)

	go func() {
		slog.Info("Starting scheduler")
		schedulder.Start()
	}()
	slog.Info("Bot starts...")
	bot.Start()
	slog.Info("App ends")
}

type Dependencies struct {
	db               *sql.DB
	telegramSubsRepo iRepo.TelegramSubscribersRepository
}

func initDependencies(dbPath string) *Dependencies {
	db, err := sqlite.InitDB(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Couldnt connect to DB. Reason: %s", err.Error()))
	}

	var telegramSubsRepo iRepo.TelegramSubscribersRepository = repositories.NewSqliteTelegramSubRepository(db)

	return &Dependencies{
		db:               db,
		telegramSubsRepo: telegramSubsRepo,
	}
}

func setupScheduledJobs(
	bot *telebot.Bot,
	telegramSubRepo iRepo.TelegramSubscribersRepository,
) gocron.Scheduler {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(fmt.Sprintf("Can't setup a cron. Reason: %s", err.Error()))
	}

	s.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(func(ctx context.Context) {
			users, err := telegramSubRepo.GetAll(ctx)
			erroredUsersCh := make(chan models.TelegramSession, len(users))
			if err != nil {
				slog.Error("Cron job error", "error", err)
			}
			g := errgroup.Group{}
			g.SetLimit(50)
			for _, user := range users {
				g.Go(func() error {
					// sleeping before send
					// this is to avoid ban from telegram
					time.Sleep(50 * time.Millisecond)
					_, err := bot.Send(&telebot.User{
						ID: user.TelegramId,
					}, "Привет чмо")
					if err != nil {
						slog.Error(fmt.Sprintf("Error sending to %d", user.TelegramId))
						erroredUsersCh <- user
						return nil
					}
					return nil
				})
			}
			g.Wait()
			close(erroredUsersCh)

			var erroredUsers []models.TelegramSession
			for user := range erroredUsersCh {
				erroredUsers = append(erroredUsers, user)
			}
			// TODO: think what to do with errors
		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	return s
}
