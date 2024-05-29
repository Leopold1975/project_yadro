package models

import "errors"

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
	RoleKey   Role = "role"
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Role         Role   `json:"role"`
	ID           int    `json:"id"`
}

type Role string

var (
	ErrNotFound      = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
)
