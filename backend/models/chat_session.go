package models

import "time"

type ChatSession struct {
    ID          int
    UserID      int
    ChatID      int
    Status      string
    StartedAt   time.Time
    FinishedAt  time.Time
    Score       int
}