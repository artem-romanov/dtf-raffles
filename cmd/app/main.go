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
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v4"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := internal.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

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
			if err != nil {
				slog.Error("Cron job error", "error", err)
			}
			// if no one is listening, exiting
			if len(users) == 0 {
				return
			}

			prevDay := time.Now().AddDate(0, 0, -1)
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
			if err := sendWithRetries(
				ctx,
				bot,
				text,
				users,
			); err != nil {
				slog.Error("Error sending raffles by schedule", "err", err)
			}

		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
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

func sendWithRetries(
	ctx context.Context,
	bot *telebot.Bot,
	message string,
	users []models.TelegramSession,
) error {
	maxRetries := 3
	maxConcurrentLimit := 10
	failed := users
	baseBackoff := 100 * time.Millisecond

	// telegram forbids sending more than 30 messages per 1 second
	// https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this
	maxSendsPerSec := 30
	limiter := rate.NewLimiter(rate.Limit(maxSendsPerSec), 1)

	for attempt := 1; attempt <= maxRetries && len(failed) > 0; attempt++ {
		attemptFailedUsers := make(chan models.TelegramSession, len(failed))
		lastAttempt := attempt == maxRetries
		g := errgroup.Group{}
		g.SetLimit(maxConcurrentLimit)

		for _, user := range failed {
			g.Go(func() error {
				if err := limiter.Wait(ctx); err != nil {
					return nil
				}

				_, err := bot.Send(
					&telebot.User{ID: user.TelegramId},
					message,
					telebot.NoPreview,
				)
				if lastAttempt && err != nil {
					slog.Error(
						fmt.Sprintf(
							"Error sending to %d, error: %v", user.TelegramId, err.Error(),
						),
					)
				}
				if err != nil {
					attemptFailedUsers <- user
					return nil
				}
				return nil
			})
		}
		g.Wait()
		close(attemptFailedUsers)

		// reconstruct failed slice
		failed = make([]models.TelegramSession, 0)
		for u := range attemptFailedUsers {
			failed = append(failed, u)
		}
		if len(failed) == 0 {
			return nil
		}

		if len(failed) > 0 && attempt < maxRetries {
			backoff := baseBackoff * time.Duration(1<<(attempt-1))
			time.Sleep(backoff)
		}
	}

	if len(failed) > 0 && len(failed) == len(users) {
		return errors.New("failed to send data to every recepient")
	}

	if len(failed) > 0 && len(failed) != len(users) {
		return errors.New("sent partially")
	}

	return nil
}
