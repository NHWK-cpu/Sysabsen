package controllers

import (
	"encoding/json"
	"net/http"

	"backend-absensi/config"
	"backend-absensi/helpers"
	"golang.org/x/crypto/bcrypt"
)

// RequestResetPassword godoc
// @Summary Minta Link Lupa Password (Guru)
// @Description Mengirim email berisi token (berlaku 15 menit)
// @Tags 1. Auth & Keamanan
// @Accept json
// @Produce json
// @Param request body controllers.ForgotPasswordReq true "Masukkan Email"
// @Success 200 {object} controllers.SuccessMessage "Link Terkirim"
// @Router /guru/forgot-password [post]

// ExecuteResetPassword godoc
// @Summary Eksekusi Reset Password
// @Description Mengganti kata sandi guru menggunakan token rahasia
// @Tags 1. Auth & Keamanan
// @Accept json
// @Produce json
// @Param request body controllers.ResetPasswordReq true "Token dan Password Baru"
// @Success 200 {object} controllers.SuccessMessage "Password berhasil diubah"
// @Router /guru/reset-password [post]
// RequestResetPassword: Menerima email dan mengirim link reset
func RequestResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Gunakan method POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	// 1. Cari user_id berdasarkan email di tabel guru
	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM guru WHERE email = ?", input.Email).Scan(&userID)
	if err != nil {
		// Walaupun tidak ketemu, kita balas sukses untuk mencegah hacker menebak email
		w.Write([]byte(`{"message": "Jika email terdaftar, link reset telah dikirim."}`))
		return
	}

	// 2. Buat Token Acak saja (Waktu kadaluarsa kita serahkan ke SQL)
	resetToken := helpers.GenerateSecureToken(32)

	// 3. Simpan Token ke tabel users dan set kadaluarsa 15 menit menggunakan DATE_ADD(NOW())
	query := "UPDATE users SET reset_token = ?, reset_expires_at = DATE_ADD(NOW(), INTERVAL 15 MINUTE) WHERE id = ?"
	_, err = config.DB.Exec(query, resetToken, userID)
	
	if err != nil {
		http.Error(w, `{"error": "Terjadi kesalahan database"}`, 500)
		return
	}

	// 4. Kirim Email
	err = helpers.SendResetEmail(input.Email, resetToken)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengirim email"}`, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Jika email terdaftar, link reset telah dikirim."}`))
}

// ExecuteResetPassword: Mengganti password
func ExecuteResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Gunakan method POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	// 1. Validasi Token & Kedaluwarsa
	// 1. Validasi Token & Kedaluwarsa secara bersamaan
	var userID int
	var isExpired bool

	// Query ini mengambil ID user, dan mengecek (True/False) apakah waktu sekarang sudah melewati batas
	query := "SELECT id, NOW() > reset_expires_at FROM users WHERE reset_token = ?"
	err := config.DB.QueryRow(query, input.Token).Scan(&userID, &isExpired)
	
	if err != nil {
		http.Error(w, `{"error": "Token tidak valid atau salah"}`, http.StatusBadRequest)
		return
	}

	if isExpired {
		http.Error(w, `{"error": "Token kedaluwarsa. Silakan minta reset ulang."}`, http.StatusForbidden)
		return
	}

	// 2. Hash Password Baru
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)

	// 3. Update Password & Bersihkan Token
	_, err = config.DB.Exec("UPDATE users SET password = ?, reset_token = NULL, reset_expires_at = NULL WHERE id = ?", string(hashedPassword), userID)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengubah password"}`, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Password berhasil diubah!"}`))
}