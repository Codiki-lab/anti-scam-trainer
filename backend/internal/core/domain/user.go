package domain

type User struct {
	ID           int
	Username     string
	PasswordHash string
	AccessRole   AccessRole
}

type AccessRole string

const (
	AccessRoleUser  AccessRole = "user"
	AccessRoleAdmin AccessRole = "admin"
)
