package models

type Link struct {
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
	IsDeleted     bool   `json:"is_deleted"`
	IsExist       *bool  `json:"is_exist"`
}
