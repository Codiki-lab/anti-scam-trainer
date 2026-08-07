package models

type User struct {
	ID             int    `json:"id" pg:"id,pk"`
	UserID         string `json:"user_id" pg:"user_id,notnull"`
	Username       string `json:"username" pg:"username,notnull"`
	CompletedChats int    `json:"completed_chats" pg:"completed_chats"`
}
