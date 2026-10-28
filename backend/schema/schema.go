package schema

import (
	"github.com/golang-jwt/jwt/v5"
)

type LoginSchema struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required"`
}

type Claims struct {
	UserID string ``
	jwt.RegisteredClaims
}
