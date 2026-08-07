package models

type Chat struct {
	ID          int    `json:"id" pg:"id,pk"`
	Title       string `json:"title" pg:"title,notnull"`
	Description string `json:"description" pg:"description"`
	Difficulty  string `json:"difficulty" pg:"difficulty,notnull"`
	Role        string `json:"role" pg:"role,notnull"`
	IsActive    bool   `json:"is_active" pg:"is_active"`
}
