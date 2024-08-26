package model

type URL struct {
	UUID        int    `json:"uuid"`
	ID          string `json:"short_url"`
	FullURL     string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// Фабричный метод для создания экземпляра URL структуры
func NewURL(id, full string) *URL {
	return &URL{
		ID:      id,
		FullURL: full,
	}
}

type APIPostRequest struct {
	URL string `json:"url"`
}

type APIPostResponse struct {
	Result string `json:"result"`
}

func NewAPIPostResponse(result string) *APIPostResponse {
	return &APIPostResponse{Result: result}
}

type APIBatchRequest struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type APIBatchResponse struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

func NewAPIBatchResponse(id, shortURL string) *APIBatchResponse {
	return &APIBatchResponse{ID: id, ShortURL: shortURL}
}

type UsersURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewUsersURLsResponse(shortURL, originalURL string) *UsersURLsResponse {
	return &UsersURLsResponse{ShortURL: shortURL, OriginalURL: originalURL}
}
