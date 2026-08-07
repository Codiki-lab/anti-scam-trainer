package domain

type Scenario struct {
	ID          int
	Title       string
	Description string
	Level       string
	UserRole    string
	IsActive    bool
}
