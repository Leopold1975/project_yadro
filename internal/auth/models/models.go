package models

import "errors"

const (
	UserRole  Role = "user"
	AdminRole Role = "admin"
	RoleKey   Role = "role"
)

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Role         Role   `json:"role"`
}

type Role string

var (
	ErrNotFound      = errors.New("user not found")
	ErrWrongPassword = errors.New("wrong password")
)
