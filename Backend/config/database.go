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
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    // Tambahkan parseTime=true untuk handle tipe data DATE/DATETIME di Go
    // Tambahkan tls=true jika Filess.io mendukung koneksi aman
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=preferred", 
        dbUser, dbPass, dbHost, dbPort, dbName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Gagal koneksi database: %v", err)
    }

    // Cek koneksi
    if err := db.Ping(); err != nil {
        log.Fatalf("Database tidak merespon: %v", err)
    }

    DB = db
    log.Println("Berhasil terhubung ke database Filess.io!")
}