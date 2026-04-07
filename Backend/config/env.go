package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// JWT_KEY kita deklarasikan di sini
var JWT_KEY []byte

// Fungsi ini akan kita panggil paling pertama saat aplikasi menyala
func LoadConfig() {
	// Membaca file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment bawaan OS")
	}

	// Mengambil rahasia dari .env dan memasukkannya ke variabel global
	JWT_KEY = []byte(os.Getenv("JWT_SECRET"))
}