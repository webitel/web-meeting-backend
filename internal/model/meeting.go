package model

type Meeting struct {
	Id        string            `json:"id" db:"id"`
	Title     string            `json:"title" db:"title"`
	CreatedAt int64             `json:"created_at" db:"created_at"`
	ExpiresAt int64             `json:"expires_at" db:"expires_at"`
	Variables map[string]string `json:"variables" db:"variables"`
	Url       string            `json:"url" db:"url"`
}
