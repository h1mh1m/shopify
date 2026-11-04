package schema

import (
	"github.com/golang-jwt/jwt/v5"
)

type LoginSchema struct {
	Email    string `form:"email" json:"email" binding:"required,email"`
	Password string `form:"password" json:"password" binding:"required"`
}

type SignUpSchema struct {
	Email        string `form:"email" json:"email" binding:"required,email"`
	Nama         string `form:"nama" json:"nama" binding:"required"`
	NomerTelepon string `form:"nomor_telepon" json:"nomor_telepon" binding:"required"`
	Password     string `form:"password" json:"password" binding:"required"`
}

type Claims struct {
	UserID string ``
	jwt.RegisteredClaims
}
