package repositories

import (
	"context"
	"database/sql"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/storage/sqlite"
	"errors"
	"fmt"
	"time"
)

const dbTableName = "telegram_subscribers"

type SqliteTelegramSubRepository struct {
	db *sql.DB
}

func NewSqliteTelegramSubRepository(db *sql.DB) *SqliteTelegramSubRepository {
	return &SqliteTelegramSubRepository{
		db: db,
	}
}

func (r *SqliteTelegramSubRepository) FindById(
	ctx context.Context,
	telegramId int64,
) (models.TelegramSession, error) {
	var user models.TelegramSession
	var createdAtRaw string
	query := fmt.Sprintf(`
		SELECT telegram_id, created_at
		FROM %s
		WHERE telegram_id = ?
		LIMIT 1;`,
		dbTableName,
	)

	row := r.db.QueryRowContext(ctx, query, telegramId)
	err := row.Scan(&user.TelegramId, &createdAtRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.TelegramSession{}, domain.ErrTelegramUserNotFound
		}
		return models.TelegramSession{}, err
	}

	user.CreatedAt, err = sqlite.FromDbTime(createdAtRaw)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *SqliteTelegramSubRepository) GetAll(ctx context.Context) ([]models.TelegramSession, error) {
	query := fmt.Sprintf(`
		SELECT telegram_id, created_at
		FROM %s;
	`, dbTableName)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.TelegramSession

	for rows.Next() {
		var session models.TelegramSession
		var createdAtRaw string

		if err = rows.Scan(&session.TelegramId, &createdAtRaw); err != nil {
			return sessions, err
		}
		if session.CreatedAt, err = sqlite.FromDbTime(createdAtRaw); err != nil {
			return sessions, err
		}

		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return sessions, err
	}

	return sessions, nil
}

func (r *SqliteTelegramSubRepository) RegisterUser(
	ctx context.Context,
	tx domain.DBTX,
	telegramId int64,
) error {
	now := time.Now()
	// check if user exists already
	_, err := r.FindById(ctx, telegramId)
	if err == nil {
		return domain.ErrTelegramUserExists
	}
	if !errors.Is(err, domain.ErrTelegramUserNotFound) {
		return err
	}

	// user not exists, lets save it then
	db := r.getExecutor(tx)
	query := fmt.Sprintf(`
		INSERT INTO %s (telegram_id, created_at) VALUES (?, ?)`,
		dbTableName,
	)
	_, err = db.ExecContext(
		ctx,
		query,
		telegramId,
		sqlite.ToDbTime(now),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *SqliteTelegramSubRepository) UnregisterUser(
	ctx context.Context,
	tx domain.DBTX,
	telegramId int64,
) error {
	// check if user exists already
	_, err := r.FindById(ctx, telegramId)
	if err != nil {
		// even if user not found - we don't care
		return fmt.Errorf("User #%d cant be unregistered: Reason: %w", telegramId, err)
	}

	db := r.getExecutor(tx)
	query := fmt.Sprintf(`
		DELETE from %s
		WHERE telegram_id = ?;
		`,
		dbTableName,
	)

	_, err = db.ExecContext(ctx, query, telegramId)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SqliteTelegramSubRepository) getExecutor(tx domain.DBTX) domain.DBTX {
	if tx != nil {
		return tx
	}
	return repo.db
}
