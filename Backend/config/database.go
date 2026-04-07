package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	// Mengambil data dari environment variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Merakit DSN secara dinamis
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Gagal menyiapkan konfigurasi database: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Gagal terhubung ke database! Error: ", err)
	}
	fmt.Println("Status: Database terhubung menggunakan konfigurasi .env!")
}