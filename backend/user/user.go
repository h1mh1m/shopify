package user

import (
	"backend/database"
	"backend/schema"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func loginUser(c *gin.Context) {
	var person schema.LoginSchema
	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user database.Customer
	err := database.DB.Where("email = ?", person.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(person.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

}

func getSecretKey() ([]byte, error) {
	key := strings.TrimSpace(os.Getenv("JWT"))
	if key == "" {
		return nil, errors.New("jwt key not configure")
	}
	if len(key) < 8 {
		return nil, errors.New("jwt key kurang dari 8 karakter")
	}
	return []byte(key), nil
}

func generateKey(userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("ID tidak ada")
	}
	secret, err := getSecretKey()
	if err != nil {
		return "", err
	}
	expireTime := time.Now().Add(2 * time.Hour)

	claim := schema.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claim)
	return token.SignedString(secret)
}

func AuthMiddlewareKey(c *gin.Context) gin.HandlerFunc {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Header dibutuhkan"})
		return nil
	}
	tokenPart := strings.SplitN(authHeader, " ", 2)
	if len(tokenPart) != 2 || tokenPart[0] != "Bearer" || strings.TrimSpace(tokenPart[1]) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "format tidak sesuai"})
		return nil
	}
	tokenString := strings.TrimSpace(tokenPart[1])

	secret, err := getSecretKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token tidak valid atau expired"})
		return nil
	}

}
