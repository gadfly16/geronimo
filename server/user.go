package server

import (
	"github.com/dgrijalva/jwt-go"
)

type User struct {
	DetailModel
	Email    string `gorm:"unique"`
	Role     string
	Password string
}

const (
	RoleUser = "user"
)

type Claims struct {
	jwt.StandardClaims
	Role string
}
