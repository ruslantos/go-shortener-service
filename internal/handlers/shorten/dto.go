package shorten

// ShortenRequest представляет структуру запроса для создания короткой ссылки.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse представляет структуру ответа для создания короткой ссылки.
type ShortenResponse struct {
	Result string `json:"result"`
}
