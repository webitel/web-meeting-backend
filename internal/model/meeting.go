package model

type Meeting struct {
	Id                string            `json:"id" db:"id"`
	DomainId          int64             `json:"domain_id" db:"domain_id"`
	Title             string            `json:"title" db:"title"`
	CreatedAt         int64             `json:"created_at" db:"created_at"`
	ExpiresAt         int64             `json:"expires_at" db:"expires_at"`
	Variables         map[string]string `json:"variables" db:"variables"`
	Url               string            `json:"url" db:"url"`
	CallId            *string           `json:"call_id" db:"call_id"`
	Satisfaction      *string           `json:"satisfaction" db:"satisfaction"`
	AllowSatisfaction bool              `json:"allow_satisfaction" db:"-"`
}
