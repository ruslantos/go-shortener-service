package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/ruslantos/go-shortener-service/internal/config"
	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/middleware/logger"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

type file interface {
	ReadEvents() ([]*fileJob.Event, error)
}

type LinksStorage struct {
	linksMap map[string]models.Links
	mutex    *sync.Mutex
	file     file
	db       *sqlx.DB
}

func NewLinksStorage(file file, db *sqlx.DB) *LinksStorage {
	return &LinksStorage{
		linksMap: make(map[string]models.Links),
		mutex:    &sync.Mutex{},
		file:     file,
		db:       db,
	}
}

func (l LinksStorage) AddLink(link models.Links) error {
	if !config.IsDatabaseExist {
		l.addLinksToMap([]models.Links{link})
		return nil
	}

	_, err := l.db.ExecContext(context.Background(),
		"INSERT INTO links  (short_url, original_url) VALUES ($1, $2)", link.ShortURL, link.OriginalURL)
	if err != nil {
		return err
	}
	return nil
}

func (l LinksStorage) AddLinkBatch(ctx context.Context, links []models.Links) error {
	if !config.IsDatabaseExist {
		l.addLinksToMap(links)
		return nil
	}

	tx, err := l.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO links (correlation_id, short_url, original_url)VALUES($1,$2,$3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range links {
		_, err = stmt.ExecContext(ctx, v.CorrelationID, v.ShortURL, v.OriginalURL)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (l LinksStorage) GetLink(value string) (string, bool, error) {
	if !config.IsDatabaseExist {
		result, ok := l.linksMap[value]
		return result.OriginalURL, ok, nil
	}

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

func (l LinksStorage) addLinksToMap(links []models.Links) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, v := range links {
		l.linksMap[v.ShortURL] = v
	}
}

func (l LinksStorage) InitLinkMap() error {
	rows, err := l.file.ReadEvents()
	if err != nil {
		return err
	}
	for _, row := range rows {
		l.linksMap[row.ShortURL] = models.Links{ShortURL: row.OriginalURL, OriginalURL: row.OriginalURL, CorrelationID: row.ID}
	}
	logger.GetLogger().Info("Links map initialized")
	return nil
}

func (l LinksStorage) Ping() error {
	if l.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	err := l.db.Ping()
	if err != nil {
		logger.GetLogger().Error(err.Error())
	}

	return err
}

func (l LinksStorage) InitDB() error {
	_, err := l.db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS links(short_url TEXT,original_url TEXT, correlation_id TEXT);`)
	if err != nil {
		logger.GetLogger().Error(err.Error())
		return err
	}

	// убрать после отказа от файла и мапы
	//for k, v := range l.linksMap {
	//	_, err := l.db.ExecContext(context.Background(),
	//		"INSERT INTO links  (short_url, original_url) VALUES ($1, $2)", k, v)
	//	if err != nil {
	//		logger.GetLogger().Error(err.Error())
	//		return err
	//	}
	//}
	return nil
}
