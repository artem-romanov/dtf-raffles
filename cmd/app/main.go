package main

import (
	"context"
	"database/sql"
	"dtf/game_draw/internal"
	"dtf/game_draw/internal/domain/models"
	iRepo "dtf/game_draw/internal/domain/repositories"
	"dtf/game_draw/internal/repositories"
	"dtf/game_draw/internal/storage/sqlite"
	"dtf/game_draw/internal/telegram"
	telegram_utils "dtf/game_draw/internal/telegram/utils"
	"dtf/game_draw/internal/usecases"
	"dtf/game_draw/pkg/dtfapi"
	"fmt"
	"log"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/telebot.v4"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := internal.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	deps := initDependencies(ctx, config.DbPath)

	bot, err := telegram.NewBot(
		config.TelegramToken,
		deps.telegramSubsRepo,
		deps.activeRafflesUseCase,
	)
	if err != nil {
		log.Fatalf("Fuck! Reason: %s", err)
	}

	schedulder := setupScheduledJobs(bot, deps.telegramSubsRepo, deps.activeRafflesUseCase)

	go func() {
		slog.Info("Starting scheduler")
		schedulder.Start()
	}()

	go func() {
		slog.Info("Bot starts...")
		bot.Start()
	}()

	<-ctx.Done()
	stop()
	slog.Info("App ends")

}

type Dependencies struct {
	db *sql.DB

	// repos
	telegramSubsRepo iRepo.TelegramSubscribersRepository
	postRepo         iRepo.PostRepository

	// usecases
	activeRafflesUseCase *usecases.GetActiveRafflePostsUseCase
}

func initDependencies(ctx context.Context, dbPath string) *Dependencies {
	db, err := sqlite.InitDB(dbPath)
	if err != nil {
		panic(fmt.Sprintf("Couldnt connect to DB. Reason: %s", err.Error()))
	}

	// external services
	dtfClient := dtfapi.NewClient(ctx)
	dtfService := dtfapi.NewService(dtfClient.Client())

	// repos
	var telegramSubsRepo iRepo.TelegramSubscribersRepository = repositories.NewSqliteTelegramSubRepository(db)
	var postRepo iRepo.PostRepository = repositories.NewDtfPostRepository(dtfService)

	// use cases
	activeRafflesUseCase := usecases.NewGetActiveRafflePostsUseCase(postRepo)

	return &Dependencies{
		db:               db,
		telegramSubsRepo: telegramSubsRepo,
		postRepo:         postRepo,

		activeRafflesUseCase: activeRafflesUseCase,
	}
}

func setupScheduledJobs(
	bot *telebot.Bot,
	telegramSubRepo iRepo.TelegramSubscribersRepository,
	activeRaffleUseCase *usecases.GetActiveRafflePostsUseCase,
) gocron.Scheduler {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic("Cant load location: " + err.Error())
	}
	s, err := gocron.NewScheduler(
		gocron.WithLocation(location),
	)
	if err != nil {
		panic(fmt.Sprintf("Can't setup a cron. Reason: %s", err.Error()))
	}

	s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(14, 0, 0),
		)),
		gocron.NewTask(func(ctx context.Context) {
			users, err := telegramSubRepo.GetAll(ctx)
			erroredUsersCh := make(chan models.TelegramSession, len(users))
			if err != nil {
				slog.Error("Cron job error", "error", err)
			}
			// if no one is listening, exiting
			if len(users) == 0 {
				return
			}

			prevDay := time.Now().AddDate(0, 0, -1)
			raffles, err := activeRaffleUseCase.Execute(ctx, prevDay)
			text := prepareTelegramText(raffles)
			if err != nil {
				// TODO: retry logic
				slog.Error("Cant load raffles from DTF", "error", err)
				return
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
					}, text, telebot.NoPreview)
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

func prepareTelegramText(posts []models.Post) string {
	return telegram_utils.ManyPostsToTelegramText(posts)
}
