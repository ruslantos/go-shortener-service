package shortenbatch

// ShortenBatchRequest представляет структуру запроса для создания нескольких коротких ссылок.
type ShortenBatchRequest []BatchOriginalURLs

// ShortenBatchResponse представляет структуру ответа для создания нескольких коротких ссылок.
type ShortenBatchResponse []BatchShortURLs

// BatchOriginalURLs представляет элемент запроса с корреляционным идентификатором и оригинальной ссылкой.
type BatchOriginalURLs struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortURLs представляет элемент ответа с корреляционным идентификатором и короткой ссылкой.
type BatchShortURLs struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
