package shorten_batch

type ShortenBatchRequest []BatchOriginalURLs
type ShortenBatchResponse []BatchShortURLs

type BatchOriginalURLs struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortURLs struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
