-- +goose Up
-- +goose StatementBegin
CREATE TABLE telegram_subscribers(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  telegram_id INTEGER UNIQUE NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE user_sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  telegram_subscriber_id INTEGER,
  access TEXT NOT NULL,
  refresh TEXT NOT NULL,
  access_expiration TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,

  FOREIGN KEY (telegram_subscriber_id)
    REFERENCES telegram_subscribers (id)
      ON UPDATE NO ACTION
      ON DELETE CASCADE
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_sessions;
DROP TABLE telegram_subscribers;
-- +goose StatementEnd
