package shorten

type ShortenRequest struct {
	Url string `json:"url"`
}
type ShortenResponse struct {
	Result string `json:"result"`
}
