package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type LinksStorage struct {
	db *sqlx.DB
}

func NewLinksStorage(db *sqlx.DB) *LinksStorage {
	return &LinksStorage{
		db: db,
	}
}

func (l LinksStorage) AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error) {
	rows, err := l.db.QueryContext(context.Background(),
		"INSERT INTO links  (short_url, original_url) VALUES ($1, $2)", link.ShortURL, link.OriginalURL)
	if err != nil || rows.Err() != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				//если url уже есть в базе, то берем из базы имеющиеся данные
				result := l.db.QueryRowContext(context.Background(),
					"SELECT short_url, original_url FROM links where original_url= $1", link.OriginalURL)
				if result.Err() != nil {
					return link, err
				}
				err = result.Scan(&link.ShortURL, &link.OriginalURL)
				if err != nil {
					return link, err
				}
				return link, internal_errors.ErrURLAlreadyExists
			}
		}
	}

	err = l.UpdateUser(ctx, link, userID)
	if err != nil {
		return link, err
	}

	return link, nil
}

func (l LinksStorage) AddLinkBatch(ctx context.Context, links []models.Link) ([]models.Link, error) {
	tx, err := l.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmtInsert, err := tx.PrepareContext(ctx,
		"INSERT INTO links (correlation_id, short_url, original_url)VALUES($1,$2,$3) "+
			"ON CONFLICT (original_url) DO NOTHING RETURNING short_url")
	if err != nil {
		return nil, err
	}
	defer stmtInsert.Close()

	stmtSelect, err := tx.PrepareContext(ctx,
		"SELECT correlation_id, short_url, original_url FROM links where original_url = $1 LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmtInsert.Close()

	var errorDB error
	for i := range links {
		v := &links[i]
		var originalURL string
		errDB := stmtInsert.QueryRowContext(ctx, v.CorrelationID, v.ShortURL, v.OriginalURL).Scan(&originalURL)
		if errDB != nil {
			if errors.Is(errDB, sql.ErrNoRows) {
				errorDB = internal_errors.ErrURLAlreadyExists
				//если url уже есть в базе, то берем из базы имеющиеся данные
				err = stmtSelect.QueryRowContext(ctx, v.OriginalURL).Scan(&v.CorrelationID, &v.ShortURL, &v.OriginalURL)
				if err != nil {
					return nil, err
				}

				continue
			}
			return nil, errDB
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return links, errorDB
}

func (l LinksStorage) GetLink(ctx context.Context, value string) (string, bool, error) {
	row := l.db.QueryRowContext(context.Background(),
		"SELECT original_url FROM links where short_url = $1 LIMIT 1", value)
	var long string
	err := row.Scan(&long)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return long, true, nil
}

func (l LinksStorage) Ping(ctx context.Context) error {
	if l.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	err := l.db.Ping()
	if err != nil {
		logger.GetLogger().Error(err.Error())
	}

	return err
}

func (l LinksStorage) InitStorage() error {
	_, err := l.db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS links(short_url TEXT,original_url TEXT, correlation_id TEXT);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON links(original_url);`)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	_, err = l.db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS users(short_url TEXT,user_id TEXT);`)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	return nil
}

func (l LinksStorage) UpdateUser(ctx context.Context, link models.Link, userID string) error {
	rows, err := l.db.QueryContext(context.Background(),
		"INSERT INTO users  (short_url, user_id) VALUES ($1, $2)", link.ShortURL, userID)

	if err != nil || rows.Err() != nil {
		return err
	}

	return nil
}

func (l LinksStorage) GetUserLinks(ctx context.Context, userID string) ([]models.Link, error) {
	var links []models.Link
	rows, err := l.db.QueryContext(ctx,
		"SELECT l.short_url, l.original_url FROM links l JOIN users u ON l.short_url = u.short_url WHERE u.user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var link models.Link
		err := rows.Scan(&link.ShortURL, &link.OriginalURL)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
