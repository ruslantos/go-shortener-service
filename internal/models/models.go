package models

// Link представляет собой структуру, содержащую информацию о короткой и оригинальной ссылках.
type Link struct {
	// ShortURL короткий идентификатор ссылки.
	ShortURL string `json:"short_url"`
	// OriginalURL оригинальная ссылка.
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
	IsDeleted     bool   `json:"is_deleted"`
	IsExist       *bool  `json:"is_exist"`
	UserID        string `json:"user_id"`
}
