package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	internal_errors "github.com/ruslantos/go-shortener-service/internal/errors"
	auth "github.com/ruslantos/go-shortener-service/internal/middleware/auth"
	"github.com/ruslantos/go-shortener-service/internal/models"
)

// MockLinksStorage реализует интерфейс LinksStorage для тестирования
type MockLinksStorage struct {
	mock.Mock
}

func (m *MockLinksStorage) AddLink(ctx context.Context, link models.Link, userID string) (models.Link, error) {
	args := m.Called(ctx, link, userID)
	return args.Get(0).(models.Link), args.Error(1)
}

func (m *MockLinksStorage) GetLink(ctx context.Context, value string) (models.Link, error) {
	args := m.Called(ctx, value)
	return args.Get(0).(models.Link), args.Error(1)
}

func (m *MockLinksStorage) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLinksStorage) AddLinkBatch(ctx context.Context, links []models.Link, userID string) ([]models.Link, error) {
	args := m.Called(ctx, links, userID)
	return args.Get(0).([]models.Link), args.Error(1)
}

func (m *MockLinksStorage) GetUserLinks(ctx context.Context, userID string) ([]models.Link, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Link), args.Error(1)
}

func (m *MockLinksStorage) DeleteUserURLs(ctx context.Context, urls []DeletedURLs) error {
	args := m.Called(ctx, urls)
	return args.Error(0)
}

func (m *MockLinksStorage) InitStorage() error {
	args := m.Called()
	return args.Error(0)
}

func TestLinkService_Get(t *testing.T) {
	tests := []struct {
		name        string
		shortLink   string
		mockSetup   func(*MockLinksStorage)
		expected    string
		expectedErr error
	}{
		{
			name:      "success",
			shortLink: "abc123",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetLink", mock.Anything, "abc123").Return(models.Link{
					ShortURL:    "abc123",
					OriginalURL: "https://example.com",
					IsDeleted:   false,
				}, nil)
			},
			expected:    "https://example.com",
			expectedErr: nil,
		},
		{
			name:      "not found",
			shortLink: "notfound",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetLink", mock.Anything, "notfound").Return(models.Link{
					IsExist:  boolPtr(false),
					ShortURL: "notfound",
				}, nil)
			},
			expected:    "notfound",
			expectedErr: internal_errors.ErrURLNotFound,
		},
		{
			name:      "deleted",
			shortLink: "deleted",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetLink", mock.Anything, "deleted").Return(models.Link{
					ShortURL:    "deleted",
					OriginalURL: "https://deleted.com",
					IsDeleted:   true,
				}, nil)
			},
			expected:    "",
			expectedErr: internal_errors.ErrURLDeleted,
		},
		{
			name:      "storage error",
			shortLink: "error",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetLink", mock.Anything, "error").Return(models.Link{}, errors.New("storage error"))
			},
			expected:    "",
			expectedErr: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockLinksStorage)
			tt.mockSetup(mockStorage)

			service := NewLinkService(mockStorage)
			result, err := service.Get(context.Background(), tt.shortLink)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestLinkService_Add(t *testing.T) {
	tests := []struct {
		name        string
		longURL     string
		userID      string
		mockSetup   func(*MockLinksStorage)
		expected    string
		expectedErr error
	}{
		{
			name:    "success",
			longURL: "https://example.com",
			userID:  "user1",
			mockSetup: func(m *MockLinksStorage) {
				m.On("AddLink", mock.Anything, mock.MatchedBy(func(link models.Link) bool {
					return link.OriginalURL == "https://example.com"
				}), "user1").Return(models.Link{
					ShortURL:    "abc123",
					OriginalURL: "https://example.com",
				}, nil)
			},
			expected:    string(mock.AnythingOfType("string")), // UUID будет сгенерирован
			expectedErr: nil,
		},
		{
			name:    "storage error",
			longURL: "https://error.com",
			userID:  "user1",
			mockSetup: func(m *MockLinksStorage) {
				m.On("AddLink", mock.Anything, mock.Anything, "user1").Return(models.Link{}, errors.New("storage error"))
			},
			expected:    "",
			expectedErr: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockLinksStorage)
			tt.mockSetup(mockStorage)

			service := NewLinkService(mockStorage)
			ctx := context.WithValue(context.Background(), auth.UserIDKey, tt.userID)
			result, err := service.Add(ctx, tt.longURL)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				if tt.name == "success" {
					// Проверяем что результат - валидный UUID
					_, err := uuid.Parse(result)
					assert.NoError(t, err)
				}
			}
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestLinkService_AddBatch(t *testing.T) {
	tests := []struct {
		name        string
		links       []models.Link
		userID      string
		mockSetup   func(*MockLinksStorage)
		expected    []models.Link
		expectedErr error
	}{
		{
			name: "success",
			links: []models.Link{
				{OriginalURL: "https://example.com/1"},
				{OriginalURL: "https://example.com/2"},
			},
			userID: "user1",
			mockSetup: func(m *MockLinksStorage) {
				m.On("AddLinkBatch", mock.Anything, mock.MatchedBy(func(links []models.Link) bool {
					return len(links) == 2 &&
						links[0].OriginalURL == "https://example.com/1" &&
						links[1].OriginalURL == "https://example.com/2"
				}), "user1").Return([]models.Link{
					{ShortURL: "abc123", OriginalURL: "https://example.com/1"},
					{ShortURL: "def456", OriginalURL: "https://example.com/2"},
				}, nil)
			},
			expected: []models.Link{
				{ShortURL: "abc123", OriginalURL: "https://example.com/1"},
				{ShortURL: "def456", OriginalURL: "https://example.com/2"},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockLinksStorage)
			tt.mockSetup(mockStorage)

			service := NewLinkService(mockStorage)
			ctx := context.WithValue(context.Background(), auth.UserIDKey, tt.userID)
			result, err := service.AddBatch(ctx, tt.links)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestLinkService_Ping(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockLinksStorage)
		expectedErr error
	}{
		{
			name: "success",
			mockSetup: func(m *MockLinksStorage) {
				m.On("Ping", mock.Anything).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "error",
			mockSetup: func(m *MockLinksStorage) {
				m.On("Ping", mock.Anything).Return(errors.New("ping error"))
			},
			expectedErr: errors.New("ping error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockLinksStorage)
			tt.mockSetup(mockStorage)

			service := NewLinkService(mockStorage)
			err := service.Ping(context.Background())

			assert.Equal(t, tt.expectedErr, err)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestLinkService_GetUserUrls(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		mockSetup   func(*MockLinksStorage)
		expected    []models.Link
		expectedErr error
	}{
		{
			name:   "success",
			userID: "user1",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetUserLinks", mock.Anything, "user1").Return([]models.Link{
					{ShortURL: "abc123", OriginalURL: "https://example.com/1"},
					{ShortURL: "def456", OriginalURL: "https://example.com/2"},
				}, nil)
			},
			expected: []models.Link{
				{ShortURL: "abc123", OriginalURL: "https://example.com/1"},
				{ShortURL: "def456", OriginalURL: "https://example.com/2"},
			},
			expectedErr: nil,
		},
		{
			name:   "empty result",
			userID: "user2",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetUserLinks", mock.Anything, "user2").Return([]models.Link{}, nil)
			},
			expected:    []models.Link{},
			expectedErr: nil,
		},
		{
			name:   "error",
			userID: "user3",
			mockSetup: func(m *MockLinksStorage) {
				m.On("GetUserLinks", mock.Anything, "user3").Return([]models.Link{}, errors.New("storage error"))
			},
			expected:    nil,
			expectedErr: errors.New("storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockLinksStorage)
			tt.mockSetup(mockStorage)

			service := NewLinkService(mockStorage)
			ctx := context.WithValue(context.Background(), auth.UserIDKey, tt.userID)
			result, err := service.GetUserUrls(ctx)

			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expectedErr, err)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestLinkService_StartDeleteWorker_ContextCancel(t *testing.T) {
	mockStorage := new(MockLinksStorage)
	service := NewLinkService(mockStorage)

	ctx, cancel := context.WithCancel(context.Background())

	// Ожидаем что DeleteUserURLs не будет вызван (так как контекст отменяется раньше)
	mockStorage.AssertNotCalled(t, "DeleteUserURLs")

	go service.StartDeleteWorker(ctx)

	// Добавляем данные для удаления
	service.ConsumeDeleteURLs(DeletedURLs{URLs: "url1", UserID: "user1"})

	// Отменяем контекст сразу
	cancel()

	// Даем время воркеру завершиться
	time.Sleep(100 * time.Millisecond)

	mockStorage.AssertExpectations(t)
}

func boolPtr(b bool) *bool {
	return &b
}
