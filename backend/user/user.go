package user

import (
	"backend/database"
	"backend/schema"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/resend/resend-go/v4"
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

func AuthMiddlewareKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Tidak dapat mengambil header"})
			return
		}
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" || tokenParts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Format tidak sesuai"})
			return
		}
		tokenString := strings.TrimSpace(tokenParts[1])
		secret, err := getSecretKey()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication Unavailable"})
			return
		}
		claims := schema.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != "ES256" {
				return nil, fmt.Errorf("Unexpected Algoritma: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token sudah expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token sudah tidak valid"})
			return
		}
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token tidak sah"})
			return
		}
		IdUser := claims.UserID
		c.Set("UserID", IdUser)

		c.Next()
	}
}

func SignUpNewUser(c *gin.Context) {
	var person schema.SignUpSchema
	if err := c.ShouldBindBodyWithJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user database.Customer
	err := database.DB.Where("email =? ", person.Email).First(&user).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already used. Please login"})
		return
	}
	HashedPassword, errB := bcrypt.GenerateFromPassword([]byte(person.Password), 12)
	if errB != nil {
		log.Fatalln("Cannot Generate password")
		return
	}

	resendAPI := os.Getenv("RESEND_API")
	secret := os.Getenv("SECRET_KEY")

	if resendAPI != "" && secret != "" {
		client := resend.NewClient(resendAPI)
		verifyLink := fmt.Sprintf("http://localhost:3000/api/verify?token=%s&email=%s", secret, person.Email)

		params := &resend.SendEmailRequest{
			From:    "Onboarding <onboarding@resend.dev>",
			To:      []string{person.Email},
			Subject: "Verifikasi Email Dev",
			Html:    fmt.Sprintf("<p>Klik link berikut untuk verifikasi: <a href='%s'>Verifikasi</a></p>", verifyLink),
		}
		if _, err := client.Emails.Send(params); err != nil {
			log.Fatalln("Gagal mengirim email verifikasi")
			return
		}

	}

}
