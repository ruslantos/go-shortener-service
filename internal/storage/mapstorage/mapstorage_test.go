package mapstorage

import (
	"context"
	"testing"

	"github.com/ruslantos/go-shortener-service/internal/models"
)

func TestGetUserLinks(t *testing.T) {
	storage := NewMapStorage()
	ctx := context.Background()
	userID := "user123"
	links := []models.Link{
		{ShortURL: "abc123", OriginalURL: "http://example.com", UserID: userID},
		{ShortURL: "def456", OriginalURL: "http://example.org", UserID: userID},
		{ShortURL: "ghi789", OriginalURL: "http://example.net", UserID: "user456"},
	}

	storage.addLinksToMap(links)

	userLinks, err := storage.GetUserLinks(ctx, userID)
	if err != nil {
		t.Errorf("GetUserLinks returned an error: %v", err)
	}

	if len(userLinks) != 2 {
		t.Errorf("GetUserLinks returned incorrect number of links: got %d, want %d", len(userLinks), 2)
	}

	for _, link := range userLinks {
		if link.UserID != userID {
			t.Errorf("GetUserLinks returned incorrect link: got %v, want UserID %s", link, userID)
		}
	}
}
