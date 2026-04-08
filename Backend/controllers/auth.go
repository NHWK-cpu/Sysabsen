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

// Login godoc
// @Summary Login Admin & Guru
// @Description Autentikasi untuk Admin dan Guru mengembalikan token JWT
// @Tags 1. Auth & Keamanan
// @Accept json
// @Produce json
// @Param request body controllers.LoginRequest true "Kredensial Login"
// @Success 200 {object} controllers.LoginSiswaResponse "Berhasil Login"
// @Router /login [post]
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

func generateTokenManual(userID int, role string) (string, error) {
	// 1. Tentukan isi "surat" (Claims)
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Berlaku 24 jam
		"iat":     time.Now().Unix(),
	}

	// 2. Buat objek token dengan metode tanda tangan HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Tanda tangani token dengan Secret Key kita
	tokenString, err := token.SignedString(config.JWT_KEY)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// LoginSiswa godoc
// @Summary Login untuk Siswa
// @Description Autentikasi siswa dan mendaftarkan ID perangkat (Device Binding)
// @Tags 1. Auth & Keamanan
// @Accept json
// @Produce json
// @Param request body controllers.LoginSiswaRequest true "Payload Data Login"
// @Success 200 {object} controllers.LoginSiswaResponse "Berhasil Login"
// @Failure 401 {object} controllers.ErrorMessage "Kredensial Salah"
// @Router /login/siswa [post]
func LoginSiswa(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username     string `json:"username"`     // Ganti Email -> Username
		Password     string `json:"password"`
		DeviceToken  string `json:"device_token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Input tidak valid", 400)
		return
	}

	// 1. Ambil data User berdasarkan USERNAME (Sesuai ERD)
	var userID int
	var hashedPassword string
	// Query diarahkan ke kolom 'username' pada tabel 'users'
	err := config.DB.QueryRow("SELECT id, password FROM users WHERE username = ? AND role = 'siswa'", input.Username).Scan(&userID, &hashedPassword)
	if err != nil {
		// Jika username tidak ditemukan
		http.Error(w, "Username atau Password salah", 401)
		return
	}

	// 2. Verifikasi Password (Bcrypt)
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		http.Error(w, "Username atau Password salah", 401)
		return
	}

	// 3. CEK DEVICE BINDING
	var status string
	err = config.DB.QueryRow("SELECT status FROM user_devices WHERE user_id = ? AND device_cookie_token = ?", userID, input.DeviceToken).Scan(&status)

	if err != nil {
		// KONDISI A: Perangkat Belum Terdaftar
		userAgent := r.Header.Get("User-Agent")
		_, insertErr := config.DB.Exec(
			"INSERT INTO user_devices (user_id, device_cookie_token, user_agent, status) VALUES (?, ?, ?, 'pending')",
			userID, input.DeviceToken, userAgent,
		)
		if insertErr != nil {
			http.Error(w, "Gagal mendaftarkan perangkat", 500)
			return
		}
		http.Error(w, "Perangkat baru terdeteksi. Harap lapor Admin untuk aktivasi login.", http.StatusForbidden)
		return
	}

	// KONDISI B: Cek Status Approval Perangkat
	if status == "pending" {
		http.Error(w, "Akses ditolak. Perangkat Anda masih menunggu persetujuan Admin.", http.StatusForbidden)
		return
	} else if status == "rejected" {
		http.Error(w, "Akses ditolak. Perangkat ini dilarang oleh Admin.", http.StatusForbidden)
		return
	}

	// KONDISI C: Approved! Buat JWT Token
	token, errToken := generateTokenManual(userID, "siswa")
	if errToken != nil {
		http.Error(w, "Gagal membuat sesi login", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Login berhasil!",
		"data": map[string]string{
			"token": token,
			"role":  "siswa",
		},
	})
}
