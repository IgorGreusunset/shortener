package storage

import (
	"context"
	"database/sql"
	"time"

	model "github.com/IgorGreusunset/shortener/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorageAdapter struct {
	DB *sql.DB
}

func NewDatabase(dbConfig string) (*DBStorageAdapter, error) {
	db, err := sql.Open("pgx", dbConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXIST shorten_urls (
		'uuid' INTEGER PRIMARY KEY,
		'short_url' VARCHAR(50),
		'original_url' TEXT,
		'created' TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return nil, err
	}

	return &DBStorageAdapter{DB: db}, nil
}

func (db *DBStorageAdapter) Create(record *model.URL) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO shorten_urls(short_url, original_url) VALUES (?, ?);`, record.ID, record.FullURL)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *DBStorageAdapter) GetByID(id string) (model.URL, bool) {
	var (
		UUID int
		ID string
		FullURL string
	)

	row := db.DB.QueryRow(`SELECT uuid, short_url, original_url FROM shorten_urls WHERE short_url = ?`, id)

	err := row.Scan(&UUID, &ID, &FullURL)
	if err != nil {
		return model.URL{}, false
	}

	result := model.NewURL(ID, FullURL)
	result.UUID = UUID
	return *result, true
}

func (db *DBStorageAdapter) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return db.DB.PingContext(ctx)
}