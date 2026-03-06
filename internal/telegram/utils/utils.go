package telegram_utils

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"
	"unicode/utf8"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v4"
)

const (
	MaxPostCharacters = 4096
)

func IsTooLongForTelegramPost(s string) bool {
	strLength := utf8.RuneCountInString(s)
	return strLength > MaxPostCharacters
}

func BroadcastWithRetries(
	ctx context.Context,
	bot telebot.API,
	message string,
	users []int64, // slice of telegram ids
) error {
	maxRetries := 3
	maxConcurrentLimit := 10
	failed := users
	baseBackoff := 100 * time.Millisecond

	// telegram forbids sending more than 30 messages per 1 second
	// https://core.telegram.org/bots/faq#my-bot-is-hitting-limits-how-do-i-avoid-this
	maxSendsPerSec := 30
	// TODO:
	// Возможно стоит вынести лимитер выше, когда посыпится много 429.
	// Пока проект для трех инвалидов - пусть будет тут, чтобы не усложнять логику
	limiter := rate.NewLimiter(rate.Limit(maxSendsPerSec), 1)

	for attempt := 1; attempt <= maxRetries && len(failed) > 0; attempt++ {
		attemptFailedUsers := make(chan int64, len(failed))
		lastAttempt := attempt == maxRetries
		g := errgroup.Group{}
		g.SetLimit(maxConcurrentLimit)

		for _, user := range failed {
			g.Go(func() error {
				if err := limiter.Wait(ctx); err != nil {
					return nil
				}

				_, err := bot.Send(
					&telebot.User{ID: user},
					message,
					telebot.NoPreview,
				)
				if lastAttempt && err != nil {
					slog.Error(
						"telegram send failed",
						"telegram_id", user,
						"error", err,
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
		failed = make([]int64, 0)
		for u := range attemptFailedUsers {
			failed = append(failed, u)
		}
		if len(failed) == 0 {
			return nil
		}

		if len(failed) > 0 && attempt < maxRetries {
			backoff := baseBackoff * time.Duration(1<<(attempt-1))
			jitterRange := backoff / 4
			jitter := time.Duration(rand.Int63n(int64(jitterRange*2))) - jitterRange
			time.Sleep(backoff + jitter)
		}
	}

	if len(failed) == len(users) {
		return errors.New("failed to send data to every recipient")
	}

	if len(failed) > 0 {
		return errors.New("sent partially")
	}

	return nil
}
