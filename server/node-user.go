package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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

func (user *User) DisplayData() (detail gin.H) {
	detail = gin.H{}
	detail["Detail"] = user
	return
}
