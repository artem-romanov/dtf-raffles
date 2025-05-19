CREATE TABLE IF NOT EXISTS user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE,  
    access TEXT NOT NULL,
    refresh TEXT NOT NULL,
    access_expiration TEXT NOT NULL,  
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
