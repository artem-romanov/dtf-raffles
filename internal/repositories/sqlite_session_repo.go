package repositories

import (
	"context"
	"database/sql"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/storage/sqlite"
	"fmt"
	"time"
)

const sqliteTableName = "user_sessions"

type SqliteUserSessionRepository struct {
	db *sql.DB
}

func NewSqliteUserSessionRepository(db *sql.DB) *SqliteUserSessionRepository {
	return &SqliteUserSessionRepository{
		db: db,
	}
}

func (repo *SqliteUserSessionRepository) GetByEmail(
	ctx context.Context,
	email string,
) (models.DtfUserSession, error) {
	var session models.DtfUserSession
	var accessExpirationString string
	queryStr := fmt.Sprintf(`
		SELECT email, access, refresh, access_expiration 
		FROM %s
		WHERE email = ?
		LIMIT 1;
	`, sqliteTableName)

	row := repo.db.QueryRowContext(ctx, queryStr, email)

	err := row.Scan(
		&session.Email,
		&session.AccessToken,
		&session.RefreshToken,
		&accessExpirationString,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return session, domain.ErrUserSessionNotFound
		}
		return session, err
	}

	session.AccessExpiration, err = sqlite.FromDbTime(accessExpirationString)

	if err != nil {
		return session, err
	}

	return session, nil
}

func (repo *SqliteUserSessionRepository) Save(
	ctx context.Context,
	tx domain.DBTX,
	us models.DtfUserSession,
) error {
	queryStr := fmt.Sprintf(`
	INSERT INTO %s (email, access, refresh, access_expiration, created_at, updated_at) 
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			access = excluded.access,
			refresh = excluded.refresh,
			access_expiration = excluded.access_expiration,
			updated_at = excluded.updated_at;
	`, sqliteTableName)

	db := repo.getExecutor(tx)
	_, err := db.ExecContext(
		ctx,
		queryStr,
		us.Email,
		us.AccessToken,
		us.RefreshToken,
		sqlite.ToDbTime(us.AccessExpiration),
		sqlite.ToDbTime(time.Now()),
		sqlite.ToDbTime(time.Now()),
	)

	if err != nil {
		return err
	}

	return nil
}

func (repo *SqliteUserSessionRepository) DeleteByEmail(
	ctx context.Context,
	tx domain.DBTX,
	email string,
) error {
	db := repo.getExecutor(tx)

	query := fmt.Sprintf(
		`DELETE FROM %s WHERE email = ?`,
		sqliteTableName,
	)

	_, err := db.ExecContext(ctx, query, email)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SqliteUserSessionRepository) getExecutor(tx domain.DBTX) domain.DBTX {
	if tx == nil {
		return repo.db
	}

	return tx
}
