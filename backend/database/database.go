package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var DB *gorm.DB

func dbConnect() *pgxpool.Pool {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Cannot Load file .env")
	}
	userDB := os.Getenv("DB_USER")
	passwordDB := os.Getenv("DB_PASSWORD")
	namaDB := os.Getenv("DB_NAME")
	hostDB := os.Getenv("DB_HOST")
	portDB := os.Getenv("DB_PORT")
	sslDB := os.Getenv("DB_SSLMODE")

	queryDB := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", userDB, passwordDB, hostDB, portDB, namaDB, sslDB)

	dbPool, err := pgxpool.New(context.Background(), queryDB)

	if err != nil {
		log.Fatalln("Tidak bisa menghubungkan ke DB")
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalln("Tidak bisa terhubung ke db")
	}
	log.Println("Bisa terhubung ke DB")

	return dbPool

}

type Customer struct {
	ID           string    `gorm:"primaryKey; column:id"`
	Nama         string    `gorm:"column:nama" json:nama`
	Email        string    `gorm:"column:email" json:email`
	NomerTelepon string    `gorm:"column:nomer_telepon" json:nomer_telepon`
	Password     string    `gorm:"column:password" json:password`
	CreatedAt    time.Time `gorm:"column:created_at" json:created_at`
	UpdateAt     time.Time `gorm:"column:update_at" json:update_at`
}
