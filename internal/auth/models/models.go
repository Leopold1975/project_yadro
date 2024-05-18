package models

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Role         Role   `json:"role"`
}

type Role string
