package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	model "github.com/IgorGreusunset/shortener/internal/app"
	"github.com/IgorGreusunset/shortener/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten_urls (
		uuid SERIAL PRIMARY KEY,
		short_url VARCHAR(50),
		original_url TEXT,
		created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return nil, err
	}

	_, err = db.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS original_url ON shorten_urls (original_url)")
	if err != nil {
		logger.Log.Debugln("Error during index creation: %v", err)
 		return nil, err
	}

	return &DBStorageAdapter{DB: db}, nil
}

func (db *DBStorageAdapter) Create(record *model.URL) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO shorten_urls(short_url, original_url, created) VALUES ($1, $2, $3);`, 
		record.ID, 
		record.FullURL, 
		time.Now())


	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr){
			if pgErr.Code == pgerrcode.UniqueViolation{
				ue := db.NewURLExistsError(record.FullURL, err)
				return ue
			}
		}
		logger.Log.Debugln(err)
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

	row := db.DB.QueryRow(`SELECT uuid, short_url, original_url FROM shorten_urls WHERE short_url = $1;`, id)

	err := row.Scan(&UUID, &ID, &FullURL)
	if err != nil {
		logger.Log.Debugln(err)
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

func (db *DBStorageAdapter) CreateBatch(urls []model.URL) error {
	tx, err := db.DB.Begin()

	if err != nil {
		return err
	}

	for _, u := range urls {
		_, err = tx.Exec(`INSERT INTO shorten_urls(short_url, original_url, created) VALUES ($1, $2, $3);`, u.ID, u.FullURL, time.Now())
		if err != nil{
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

type URLExistsError struct {
	ShortURL string
	Er string
}

func (db *DBStorageAdapter) NewURLExistsError(originalURL string, e error) *URLExistsError {
	var ID string
	row := db.DB.QueryRow(`SELECT short_url FROM shorten_urls WHERE original_url = $1;`, originalURL)
	row.Scan(&ID)
	return &URLExistsError{ShortURL: ID, Er: "Original URL already in DB"}
}

func (uee *URLExistsError) Error() string {
	return uee.Er
}