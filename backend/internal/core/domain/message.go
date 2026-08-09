package domain

import "time"

// Message is a visible dialogue line produced by the user or virtual interlocutor.
type Message struct {
	ID        int
	AttemptID int
	Author    MessageAuthor
	Text      string
	CreatedAt time.Time
}

type MessageAuthor string

const (
	MessageAuthorUser         MessageAuthor = "user"
	MessageAuthorInterlocutor MessageAuthor = "interlocutor"
)
