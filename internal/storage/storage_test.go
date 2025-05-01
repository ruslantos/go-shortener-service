package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	"github.com/ruslantos/go-shortener-service/internal/models"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

func TestAddLink(t *testing.T) {
	tests := []struct {
		name        string
		link        models.Link
		userID      string
		mock        func(mock sqlmock.Sqlmock)
		expected    models.Link
		expectedErr error
	}{
		{
			name:   "successful add",
			link:   models.Link{ShortURL: "abc", OriginalURL: "http://example.com"},
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO links").
					WithArgs("abc", "http://example.com", "user1").
					WillReturnRows(sqlmock.NewRows([]string{}))
			},
			expected:    models.Link{ShortURL: "abc", OriginalURL: "http://example.com"},
			expectedErr: nil,
		},
		{
			name:   "duplicate url",
			link:   models.Link{ShortURL: "abc", OriginalURL: "http://example.com"},
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO links").
					WithArgs("abc", "http://example.com", "user1").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				mock.ExpectQuery("SELECT short_url, original_url FROM links where original_url= ?").
					WithArgs("http://example.com").
					WillReturnRows(sqlmock.NewRows([]string{"short_url", "original_url"}).
						AddRow("def", "http://example.com"))
			},
			expected:    models.Link{ShortURL: "def", OriginalURL: "http://example.com"},
			expectedErr: internal_errors.ErrURLAlreadyExists,
		},
		{
			name:   "other database error",
			link:   models.Link{ShortURL: "abc", OriginalURL: "http://example.com"},
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO links").
					WithArgs("abc", "http://example.com", "user1").
					WillReturnError(errors.New("database error"))
			},
			expected:    models.Link{ShortURL: "abc", OriginalURL: "http://example.com"},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			storage := NewLinksStorage(sqlxDB)

			tt.mock(mock)

			result, err := storage.AddLink(context.Background(), tt.link, tt.userID)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestAddLinkBatch(t *testing.T) {
	tests := []struct {
		name        string
		links       []models.Link
		userID      string
		mock        func(mock sqlmock.Sqlmock)
		expected    []models.Link
		expectedErr error
	}{
		{
			name: "successful batch add",
			links: []models.Link{
				{CorrelationID: "1", ShortURL: "abc", OriginalURL: "http://example.com"},
			},
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO links")
				mock.ExpectPrepare("SELECT correlation_id, short_url, original_url FROM links")

				mock.ExpectQuery("INSERT INTO links").
					WithArgs("1", "abc", "http://example.com", "user1").
					WillReturnRows(sqlmock.NewRows([]string{"short_url"}).AddRow("abc"))

				mock.ExpectCommit()
			},
			expected: []models.Link{
				{CorrelationID: "1", ShortURL: "abc", OriginalURL: "http://example.com"},
			},
			expectedErr: nil,
		},
		{
			name: "duplicate url in batch",
			links: []models.Link{
				{CorrelationID: "1", ShortURL: "abc", OriginalURL: "http://example.com"},
			},
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO links")
				mock.ExpectPrepare("SELECT correlation_id, short_url, original_url FROM links")

				mock.ExpectQuery("INSERT INTO links").
					WithArgs("1", "abc", "http://example.com", "user1").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery("SELECT correlation_id, short_url, original_url FROM links where original_url = ?").
					WithArgs("http://example.com").
					WillReturnRows(sqlmock.NewRows([]string{"correlation_id", "short_url", "original_url"}).
						AddRow("1", "def", "http://example.com"))

				mock.ExpectCommit()
			},
			expected: []models.Link{
				{CorrelationID: "1", ShortURL: "def", OriginalURL: "http://example.com"},
			},
			expectedErr: internal_errors.ErrURLAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			storage := NewLinksStorage(sqlxDB)

			tt.mock(mock)

			result, err := storage.AddLinkBatch(context.Background(), tt.links, tt.userID)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetLink(t *testing.T) {
	tests := []struct {
		name        string
		shortURL    string
		mock        func(mock sqlmock.Sqlmock)
		expected    models.Link
		expectedErr error
	}{
		{
			name:     "successful get",
			shortURL: "abc",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT original_url, is_deleted FROM links where short_url = ?").
					WithArgs("abc").
					WillReturnRows(sqlmock.NewRows([]string{"original_url", "is_deleted"}).
						AddRow("http://example.com", false))
			},
			expected: models.Link{
				OriginalURL: "http://example.com",
				IsDeleted:   false,
			},
			expectedErr: nil,
		},
		{
			name:     "not found",
			shortURL: "abc",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT original_url, is_deleted FROM links where short_url = ?").
					WithArgs("abc").
					WillReturnError(sql.ErrNoRows)
			},
			expected: models.Link{
				IsExist: new(bool),
			},
			expectedErr: nil,
		},
		{
			name:     "deleted link",
			shortURL: "abc",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT original_url, is_deleted FROM links where short_url = ?").
					WithArgs("abc").
					WillReturnRows(sqlmock.NewRows([]string{"original_url", "is_deleted"}).
						AddRow("http://example.com", true))
			},
			expected: models.Link{
				OriginalURL: "http://example.com",
				IsDeleted:   true,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			storage := NewLinksStorage(sqlxDB)

			tt.mock(mock)

			result, err := storage.GetLink(context.Background(), tt.shortURL)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetUserLinks(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		mock        func(mock sqlmock.Sqlmock)
		expected    []models.Link
		expectedErr error
	}{
		{
			name:   "successful get user links",
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"short_url", "original_url"}).
					AddRow("abc", "http://example.com").
					AddRow("def", "http://example.org")
				mock.ExpectQuery("SELECT short_url, original_url FROM links WHERE user_id = ?").
					WithArgs("user1").
					WillReturnRows(rows)
			},
			expected: []models.Link{
				{ShortURL: "abc", OriginalURL: "http://example.com"},
				{ShortURL: "def", OriginalURL: "http://example.org"},
			},
			expectedErr: nil,
		},
		{
			name:   "no links for user",
			userID: "user1",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT short_url, original_url FROM links WHERE user_id = ?").
					WithArgs("user1").
					WillReturnRows(sqlmock.NewRows([]string{"short_url", "original_url"}))
			},
			expected:    nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			storage := NewLinksStorage(sqlxDB)

			tt.mock(mock)

			result, err := storage.GetUserLinks(context.Background(), tt.userID)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDeleteUserURLs(t *testing.T) {
	tests := []struct {
		name        string
		urls        []service.DeletedURLs
		mock        func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "empty urls",
			urls: []service.DeletedURLs{},
			mock: func(mock sqlmock.Sqlmock) {
				// No expectations as the method should return early
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			storage := NewLinksStorage(sqlxDB)

			tt.mock(mock)

			err = storage.DeleteUserURLs(context.Background(), tt.urls)
			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
