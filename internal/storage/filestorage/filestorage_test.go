package filestorage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	fileJob "github.com/ruslantos/go-shortener-service/internal/files"
	"github.com/ruslantos/go-shortener-service/internal/models"
	"github.com/ruslantos/go-shortener-service/internal/service"
)

// MockFileConsumer реализует FileConsumer для тестов
type MockFileConsumer struct {
	mock.Mock
}

func (m *MockFileConsumer) ReadEvents() ([]*fileJob.Event, error) {
	args := m.Called()
	return args.Get(0).([]*fileJob.Event), args.Error(1)
}

// MockFileProducer реализует FileProducer для тестов
type MockFileProducer struct {
	mock.Mock
}

func (m *MockFileProducer) WriteEvent(event *fileJob.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestNewFileStorage(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}

	storage := NewFileStorage(consumer, producer)

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.linksMap)
	assert.NotNil(t, storage.mutex)
	assert.Equal(t, consumer, storage.fileConsumer)
	assert.Equal(t, producer, storage.fileProducer)
}

func TestAddLink(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	link := models.Link{
		ShortURL:    "abc",
		OriginalURL: "http://example.com",
		UserID:      "user1",
	}

	producer.On("WriteEvent", &fileJob.Event{
		ID:          link.CorrelationID,
		ShortURL:    link.ShortURL,
		OriginalURL: link.OriginalURL,
	}).Return(nil)

	result, err := storage.AddLink(context.Background(), link, "user1")

	assert.NoError(t, err)
	assert.Equal(t, link, result)
	assert.Equal(t, link, storage.linksMap["abc"])
	producer.AssertExpectations(t)
}

func TestAddLinkBatch(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	links := []models.Link{
		{ShortURL: "abc", OriginalURL: "http://example.com", UserID: "user1"},
		{ShortURL: "def", OriginalURL: "http://example.org", UserID: "user1"},
	}

	for _, link := range links {
		producer.On("WriteEvent", &fileJob.Event{
			ID:          link.CorrelationID,
			ShortURL:    link.ShortURL,
			OriginalURL: link.OriginalURL,
		}).Return(nil)
	}

	result, err := storage.AddLinkBatch(context.Background(), links, "user1")

	assert.NoError(t, err)
	assert.Equal(t, links, result)
	assert.Equal(t, links[0], storage.linksMap["abc"])
	assert.Equal(t, links[1], storage.linksMap["def"])
	producer.AssertExpectations(t)
}

func TestGetLink(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	expectedLink := models.Link{
		ShortURL:    "abc",
		OriginalURL: "http://example.com",
		UserID:      "user1",
	}
	storage.linksMap["abc"] = expectedLink

	link, err := storage.GetLink(context.Background(), "abc")

	assert.NoError(t, err)
	assert.Equal(t, expectedLink.OriginalURL, link.OriginalURL)
}

func TestGetLink_NotFound(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	link, err := storage.GetLink(context.Background(), "nonexistent")

	assert.NoError(t, err)
	assert.Equal(t, "", link.OriginalURL)
}

func TestInitStorage(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	events := []*fileJob.Event{
		{ID: "1", ShortURL: "abc", OriginalURL: "http://example.com"},
		{ID: "2", ShortURL: "def", OriginalURL: "http://example.org"},
	}

	consumer.On("ReadEvents").Return(events, nil)

	err := storage.InitStorage()

	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", storage.linksMap["abc"].OriginalURL)
	assert.Equal(t, "http://example.org", storage.linksMap["def"].OriginalURL)
	consumer.AssertExpectations(t)
}

func TestInitStorage_Error(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	expectedErr := errors.New("read error")
	consumer.On("ReadEvents").Return([]*fileJob.Event{}, expectedErr)

	err := storage.InitStorage()

	assert.Equal(t, expectedErr, err)
	consumer.AssertExpectations(t)
}

func TestGetUserLinks(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	links := []models.Link{
		{ShortURL: "abc", OriginalURL: "http://example.com", UserID: "user1"},
		{ShortURL: "def", OriginalURL: "http://example.org", UserID: "user2"},
		{ShortURL: "ghi", OriginalURL: "http://example.net", UserID: "user1"},
	}

	for _, link := range links {
		storage.linksMap[link.ShortURL] = link
	}

	userLinks, err := storage.GetUserLinks(context.Background(), "user1")

	assert.NoError(t, err)
	assert.Len(t, userLinks, 2)
	assert.Contains(t, userLinks, links[0])
	assert.Contains(t, userLinks, links[2])
}

func TestDeleteUserURLs(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	link := models.Link{
		ShortURL:    "abc",
		OriginalURL: "http://example.com",
		UserID:      "user1",
		IsDeleted:   false,
	}
	storage.linksMap["abc"] = link

	urls := []service.DeletedURLs{
		{UserID: "user1", URLs: "abc"},
	}

	err := storage.DeleteUserURLs(context.Background(), urls)

	assert.NoError(t, err)
	assert.True(t, storage.linksMap["abc"].IsDeleted)
}

func TestDeleteUserURLs_NotFound(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	urls := []service.DeletedURLs{
		{UserID: "user1", URLs: "nonexistent"},
	}

	err := storage.DeleteUserURLs(context.Background(), urls)

	assert.Error(t, err)
	assert.Equal(t, "url not found in storage", err.Error())
}

func TestPing(t *testing.T) {
	consumer := &MockFileConsumer{}
	producer := &MockFileProducer{}
	storage := NewFileStorage(consumer, producer)

	err := storage.Ping(context.Background())

	assert.NoError(t, err)
}
