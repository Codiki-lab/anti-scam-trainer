package domain

type User struct {
	ID             int
	ExternalID     string
	Username       string
	CompletedChats int
}
