package domain

type User struct {
	ID           int
	Username     string
	PasswordHash string
	AccessRole   AccessRole
	TrainingRole UserRole
	Streak       Streak
}

type UserRole string

const (
	UserRoleBuyer  UserRole = "buyer"
	UserRoleSeller UserRole = "seller"
)

func ValidUserRole(role UserRole) bool { return role == UserRoleBuyer || role == UserRoleSeller }

type Streak struct {
	Current          int    `json:"current"`
	Longest          int    `json:"longest"`
	ActiveToday      bool   `json:"active_today"`
	LastActivityDate string `json:"last_activity_date,omitempty"`
}

type AccessRole string

const (
	AccessRoleUser  AccessRole = "user"
	AccessRoleAdmin AccessRole = "admin"
)
