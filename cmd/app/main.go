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
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gopkg.in/telebot.v4"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := internal.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	initSlog()

	deps, cleanup := initDependencies(ctx, config.DbPath)
	defer func() {
		if err := cleanup(); err != nil {
			slog.Error("dependencies cleanup error", "err", err)
		}
	}()

	bot, err := telegram.NewBot(
		config.TelegramToken,
		deps.telegramSubsRepo,
		deps.activeRafflesUseCase,
		config.TelegramAdmins,
	)
	if err != nil {
		log.Fatalf("Fuck! Reason: %s", err)
	}

	schedulder := setupScheduledJobs(bot, deps.telegramSubsRepo, deps.activeRafflesUseCase)
	defer schedulder.Shutdown()

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

func initDependencies(ctx context.Context, dbPath string) (*Dependencies, func() error) {
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

	// function to clean all generated shit
	cleanup := func() error {
		var errs []error
		if err := db.Close(); err != nil {
			errs = append(errs, err)
		}

		if err := dtfClient.Close(); err != nil {
			errs = append(errs, err)
		}

		if len(errs) != 0 {
			return errors.Join(errs...)
		}

		return nil
	}

	return &Dependencies{
		db:               db,
		telegramSubsRepo: telegramSubsRepo,
		postRepo:         postRepo,

		activeRafflesUseCase: activeRafflesUseCase,
	}, cleanup
}

func initSlog() {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
	)
	slog.SetDefault(logger)
}

func setupScheduledJobs(
	bot *telebot.Bot,
	telegramSubRepo iRepo.TelegramSubscribersRepository,
	activeRaffleUseCase *usecases.GetActiveRafflePostsUseCase,
) gocron.Scheduler {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("Cant load location: %v\n", err)
	}
	s, err := gocron.NewScheduler(
		gocron.WithLocation(location),
	)
	if err != nil {
		log.Fatalf("Can't setup a cron. Reason: %v\n", err)
	}

	_, err = s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(14, 0, 0),
		)),
		gocron.NewTask(func(ctx context.Context) {
			users, err := telegramSubRepo.GetAll(ctx)
			if err != nil {
				slog.Error("Cron job error", "error", err)
			}
			// if no one is listening, exiting
			if len(users) == 0 {
				return
			}

			prevDay := time.Now().In(location).AddDate(0, 0, -1)
			raffles, err := activeRaffleUseCase.Execute(ctx, prevDay)
			if err != nil {
				// TODO: retry logic
				slog.Error("Cant load raffles from DTF", "error", err)
				return
			}
			// dont bother users when no active raffles
			if len(raffles) == 0 {
				return
			}

			// ok, lets send this shit to users
			text := prepareTelegramText(raffles)
			if err := telegram_utils.BroadcastWithRetries(
				ctx,
				bot,
				text,
				models.TelegramSessionsToIds(users),
			); err != nil {
				slog.Error("Error sending raffles by schedule", "err", err)
			}

		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		slog.Error("couldn't setup scheduled job", "err", err)
	}
	return s
}

func prepareTelegramText(posts []models.Post) string {
	text := telegram_utils.ManyPostsToTelegramText(posts, false)
	if telegram_utils.IsTooLongForTelegramPost(text) {
		// do the shorten version
		text = telegram_utils.ManyPostsToTelegramText(posts, true)
	}
	return text
}
