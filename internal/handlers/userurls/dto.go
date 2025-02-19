package userurls

type UserURLsResponse []UserURLs

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
