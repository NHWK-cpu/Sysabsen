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
	// PERUBAHAN: Menambahkan AND is_active = 1
	// Hanya user yang aktif yang bisa ditemukan oleh query ini
	// ==========================================
	err := config.DB.QueryRow("SELECT id, password, role FROM users WHERE username = ? AND is_active = 1", req.Username).Scan(&userId, &dbPassword, &userRole)

	if err != nil {
		// Pesan error diubah sedikit agar lebih informatif
		http.Error(w, `{"error": "Username/password salah atau akun telah dinonaktifkan"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(req.Password)); err != nil {
		http.Error(w, `{"error": "Username atau password salah"}`, http.StatusUnauthorized)
		return
	}

	config.DB.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", userId)

	var namaUser string
	if userRole == "guru" {
		// Catatan: Sesuaikan 'nama_guru' dengan nama kolom di tabel 'guru' milikmu
		errGuru := config.DB.QueryRow("SELECT nama_lengkap FROM guru WHERE user_id = ?", userId).Scan(&namaUser)
		if errGuru != nil {
			namaUser = req.Username // Fallback jika tidak ketemu
		}
	} else {
		namaUser = req.Username // Jika yang login adalah admin
	}

	claims := jwt.MapClaims{
		"user_id":  userId,
		"role":     userRole,
		"username": req.Username,
		"nama":     namaUser,
		"exp":      time.Now().Add(time.Minute * 90).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(config.JWT_KEY)
	if err != nil {
		http.Error(w, `{"error": "Gagal membuat token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{"message": "Login sukses!", "role": "%s", "token": "%s"}`, userRole, tokenString)
	w.Write([]byte(response))
}

func generateTokenManual(userID int, role string, username string, namaLengkap string, namaSekolah string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":      userID,
		"role":         role,
		"username":     username,
		"nama_lengkap": namaLengkap, // Sinkron dengan pembacaan decodedPayload.nama_lengkap
		"nama_sekolah": namaSekolah, // Ini yang dicari oleh dashboard frontend!
		"exp":          time.Now().Add(60 * time.Minute).Unix(),
		"iat":          time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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
		Username    string `json:"username"`
		Password    string `json:"password"`
		DeviceToken string `json:"device_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Input tidak valid", 400)
		return
	}

	var userID int
	var hashedPassword string

	err := config.DB.QueryRow("SELECT id, password FROM users WHERE username = ? AND role = 'siswa' AND is_active = 1", input.Username).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "Username/Password salah atau akun Anda telah dinonaktifkan", 401)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password))
	if err != nil {
		http.Error(w, "Username atau Password salah", 401)
		return
	}

	var status string
	err = config.DB.QueryRow("SELECT status FROM user_devices WHERE user_id = ? AND device_cookie_token = ?", userID, input.DeviceToken).Scan(&status)

	if err != nil {
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

	if status == "pending" {
		http.Error(w, "Akses ditolak. Perangkat Anda masih menunggu persetujuan Admin.", http.StatusForbidden)
		return
	} else if status == "rejected" {
		http.Error(w, "Akses ditolak. Perangkat ini dilarang oleh Admin.", http.StatusForbidden)
		return
	}

	config.DB.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", userID)
	
	var namaSiswa, namaSekolah string
	
	// PERUBAHAN: Tarik sekalian nama_sekolah dari tabel siswa
	errQuery := config.DB.QueryRow("SELECT nama_lengkap, nama_sekolah FROM siswa WHERE user_id = ?", userID).Scan(&namaSiswa, &namaSekolah)
	
	if errQuery != nil {
		namaSiswa = input.Username
		
		// Fallback anti-badai: Cari di tabel users jika ternyata disimpannya di kolom identifier
		var identifier string
		errFallback := config.DB.QueryRow("SELECT identifier FROM users WHERE id = ?", userID).Scan(&identifier)
		if errFallback == nil && identifier != "" {
			namaSekolah = identifier
		} else {
			namaSekolah = "Asal Sekolah Tidak Diketahui"
		}
	}

	// PERUBAHAN: Kirim variabel namaSekolah ke fungsi pembuat token
	token, errToken := generateTokenManual(userID, "siswa", input.Username, namaSiswa, namaSekolah)
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
