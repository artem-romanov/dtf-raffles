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
	"dtf/game_draw/internal/usecases"
	"dtf/game_draw/pkg/dtfapi"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"golang.org/x/sync/errgroup"
	"gopkg.in/telebot.v4"
)

func main() {
	ctx := context.Background()
	config, err := internal.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	fmt.Println(config)

	deps := initDependencies(ctx, config.DbPath)

	bot, err := telegram.NewBot(config.TelegramToken, deps.telegramSubsRepo)
	if err != nil {
		log.Fatalf("Fuck! Reason: %s", err)
	}

	schedulder := setupScheduledJobs(bot, deps.telegramSubsRepo, deps.activeRafflesUseCase)

	go func() {
		slog.Info("Starting scheduler")
		schedulder.Start()
	}()
	slog.Info("Bot starts...")
	bot.Start()
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
			// if no one is listening, exiting
			if len(users) == 0 {
				return
			}

			prevDay := time.Now().AddDate(0, 0, -1)
			raffles, err := activeRaffleUseCase.Execute(prevDay)
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
					}, text)
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

func prettyRaffle(post models.Post) string {
	header := fmt.Sprintf("Новость %d: %s\n", post.Id, post.Title)
	link := fmt.Sprintf("Ссылка: %s", post.Uri)

	return header + link
}

func prepareTelegramText(posts []models.Post) string {
	builder := strings.Builder{}
	for i, post := range posts {
		if i > 0 && i < len(posts) {
			builder.WriteString("\n * * * \n")
		}
		text := prettyRaffle(post)
		builder.WriteString(text)
	}
	return builder.String()
}
