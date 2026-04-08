package controllers

import (
	"encoding/json"
	"net/http"

	"backend-absensi/config"

	"golang.org/x/crypto/bcrypt"
)

// GetPendingDevices godoc
// @Summary Daftar Device Menunggu Approval
// @Description Melihat daftar HP siswa yang menunggu persetujuan login
// @Tags 4. Admin - Device & Keamanan
// @Produce json
// @Security BearerAuth
// @Success 200 {array} controllers.DevicePendingList "Daftar Device"
// @Router /admin/device/pending [get]

// ApproveDevice godoc
// @Summary Setujui Perangkat Siswa
// @Tags 4. Admin - Device & Keamanan
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body controllers.DeviceActionReq true "ID Perangkat"
// @Success 200 {object} controllers.SuccessMessage "Perangkat Disetujui"
// @Router /admin/device/approve [post]

// ResetPasswordSiswa godoc
// @Summary Reset Password Siswa (Lupa Password)
// @Description Admin mereset password siswa dengan validasi Kata Kunci Rahasia
// @Tags 4. Admin - Device & Keamanan
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body controllers.ResetPassSiswaReq true "Data Reset"
// @Success 200 {object} controllers.SuccessMessage "Password berhasil direset"
// @Router /admin/siswa/reset-password [post]
// GetPendingDevices: Mengambil daftar perangkat yang statusnya 'pending'
func GetPendingDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Gunakan method GET", http.StatusMethodNotAllowed)
		return
	}

	// Query JOIN untuk mendapatkan nama siswa yang memiliki perangkat tersebut
	query := `
		SELECT ud.id, s.nama_lengkap, ud.device_cookie_token, ud.user_agent, ud.created_at 
		FROM user_devices ud
		JOIN siswa s ON ud.user_id = s.user_id
		WHERE ud.status = 'pending'
		ORDER BY ud.created_at ASC`

	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, "Gagal mengambil data perangkat: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type PendingDevice struct {
		ID                int    `json:"id"`
		NamaSiswa         string `json:"nama_siswa"`
		DeviceCookieToken string `json:"device_cookie_token"`
		UserAgent         string `json:"user_agent"`
		CreatedAt         string `json:"created_at"`
	}

	var devices []PendingDevice
	for rows.Next() {
		var d PendingDevice
		rows.Scan(&d.ID, &d.NamaSiswa, &d.DeviceCookieToken, &d.UserAgent, &d.CreatedAt)
		devices = append(devices, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// ApproveDevice: Mengubah status perangkat menjadi 'approved'
func ApproveDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Gunakan method POST", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		DeviceID int `json:"device_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	// Update status menjadi 'approved'
	_, err := config.DB.Exec("UPDATE user_devices SET status = 'approved' WHERE id = ?", input.DeviceID)
	if err != nil {
		http.Error(w, "Gagal meng-approve perangkat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sip! Perangkat telah berhasil di-approve."))
}

// RejectDevice godoc
// @Summary Tolak Perangkat Siswa
// @Description Menolak akses login dari HP yang tidak dikenal
// @Tags 4. Admin - Dashboard & Device
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body controllers.DeviceActionReq true "ID Perangkat"
// @Success 200 {object} controllers.SuccessMessage "Perangkat Ditolak"
// @Router /admin/device/reject [post]
// RejectDevice: Mengubah status perangkat menjadi 'rejected'
func RejectDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Gunakan method POST", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		DeviceID int `json:"device_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	// Update status menjadi 'rejected'
	// Pastikan kata 'rejected' sama persis dengan yang ada di ENUM database
	_, err := config.DB.Exec("UPDATE user_devices SET status = 'rejected' WHERE id = ?", input.DeviceID)
	if err != nil {
		http.Error(w, "Gagal menolak perangkat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Perangkat telah berhasil ditolak. Siswa tidak akan bisa login dari perangkat ini."))
}

// ResetPasswordSiswa: Admin mereset password setelah memverifikasi kata_kunci
func ResetPasswordSiswa(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NIS         string `json:"nis"`
		KataKunci   string `json:"kata_kunci"`
		NewPassword string `json:"new_password"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	// 1. Verifikasi apakah NIS dan Kata Kunci cocok
	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM siswa WHERE nis = ? AND kata_kunci = ?", input.NIS, input.KataKunci).Scan(&userID)
	if err != nil {
		http.Error(w, "NIS atau Kata Kunci salah. Verifikasi gagal.", 403)
		return
	}

	// 2. Hash Password Baru
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)

	// 3. Update di tabel users
	_, err = config.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashedPassword), userID)
	if err != nil {
		http.Error(w, "Gagal mengupdate password", 500)
		return
	}

	w.Write([]byte("Password siswa berhasil direset!"))
}