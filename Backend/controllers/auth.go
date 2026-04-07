package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend-absensi/config"
	"backend-absensi/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Hanya menerima method POST", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format data salah", http.StatusBadRequest)
		return
	}

	var dbPassword string
	var userId int
	var userRole string
	
	// ==========================================
	// PERUBAHAN: Kita hapus "AND role = 'admin'"
	// Sekarang query ini mencari user siapapun itu (Admin/Guru/Siswa)
	// ==========================================
	err := config.DB.QueryRow("SELECT id, password, role FROM users WHERE username = ?", req.Username).Scan(&userId, &dbPassword, &userRole)

	if err != nil {
		http.Error(w, `{"error": "Username atau password salah"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "Username atau password salah"}`, http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"user_id": userId,
		"role":    userRole, // Role siswa akan otomatis tercatat di dalam tiket ini
		"exp":     time.Now().Add(time.Hour * 24).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(config.JWT_KEY)
	if err != nil {
		http.Error(w, `{"error": "Gagal membuat token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// PERUBAHAN KEDUA: Kita kembalikan informasi 'role' ke klien agar Frontend
	// tahu apakah harus membuka halaman dashboard admin, atau halaman absen siswa.
	response := fmt.Sprintf(`{"message": "Login sukses!", "role": "%s", "token": "%s"}`, userRole, tokenString)
	w.Write([]byte(response))
}