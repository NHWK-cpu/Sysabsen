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
// RequestResetPassword godoc
// @Summary Minta Link Lupa Password (Guru)
// @Description Mengirim email berisi token menggunakan Username
func RequestResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Gunakan method POST"}`, http.StatusMethodNotAllowed)
		return
	}

	// 1. Ubah input untuk menerima Username
	var input struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Format request tidak valid"}`, http.StatusBadRequest)
		return
	}

	var userID int
	var targetEmail string

	// 2. Kueri JOIN: Cari user_id dan email berdasarkan username dari tabel users dan guru
	querySelect := `
	SELECT u.id, g.email
	FROM users u
	JOIN guru g ON u.id = g.user_id
	WHERE u.username = ? AND u.role = 'guru' AND u.is_active = 1
	`
	err := config.DB.QueryRow(querySelect, input.Username).Scan(&userID, &targetEmail)

	if err != nil {
		// Balas dengan pesan sukses palsu (Security Best Practice) agar hacker tidak bisa menebak username mana yang valid
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Jika username valid dan memiliki email, link reset telah dikirim."}`))
		return
	}

	// Cek ekstra: Pastikan email tidak kosong (berjaga-jaga jika ada guru yang datanya belum lengkap)
	if targetEmail == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Jika username valid dan memiliki email, link reset telah dikirim."}`))
		return
	}

	// 3. Buat Token Acak
	resetToken := helpers.GenerateSecureToken(32)

	// 4. Simpan Token ke tabel users
	queryUpdate := "UPDATE users SET reset_token = ?, reset_expires_at = DATE_ADD(NOW(), INTERVAL 15 MINUTE) WHERE id = ?"
	_, err = config.DB.Exec(queryUpdate, resetToken, userID)

	if err != nil {
		http.Error(w, `{"error": "Terjadi kesalahan sistem internal"}`, 500)
		return
	}

	// 5. Kirim Email ke targetEmail yang didapat dari tabel guru
	err = helpers.SendResetEmail(targetEmail, resetToken)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengirim email reset"}`, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Jika username valid dan memiliki email, link reset telah dikirim."}`))
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
