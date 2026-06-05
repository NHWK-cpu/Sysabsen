package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"backend-absensi/config"
	"backend-absensi/models"

	"golang.org/x/crypto/bcrypt"
)

// Fitur CREATE
func CreateSiswa(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Hanya menerima method POST", http.StatusMethodNotAllowed)
        return
    }

    var req models.SiswaRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengenkripsi password"}`, http.StatusInternalServerError)
        return
    }

    tx, err := config.DB.Begin()
    if err != nil {
        http.Error(w, `{"error": "Gagal memulai transaksi database"}`, http.StatusInternalServerError)
        return
    }

    result, err := tx.Exec("INSERT INTO users (username, password, role, is_active) VALUES (?, ?, 'siswa', 1)", req.Username, hashedPassword)
    if err != nil {
        tx.Rollback()
        http.Error(w, `{"error": "Username mungkin sudah dipakai"}`, http.StatusConflict)
        return
    }

    userID, _ := result.LastInsertId()

    _, err = tx.Exec("INSERT INTO siswa (user_id, nama_sekolah, nama_lengkap, label_kata_kunci, kata_kunci) VALUES (?, ?, ?, ?, ?)", 
        userID, req.NamaSekolah, req.NamaLengkap, req.LabelKataKunci, req.KataKunci)
    
    if err != nil {
        tx.Rollback() 
        http.Error(w, `{"error": "Gagal menyimpan data profil siswa ke database."}`, http.StatusInternalServerError)
        return
    }

    tx.Commit()

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(fmt.Sprintf(`{"message": "Siswa %s berhasil ditambahkan!"}`, req.NamaLengkap)))
}

// RegisterSiswaPublic — daftar mandiri; akun nonaktif sampai admin menyetujui
func RegisterSiswaPublic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Hanya menerima method POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.SiswaSelfRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.NamaSekolah = strings.TrimSpace(req.NamaSekolah)
	req.NamaLengkap = strings.TrimSpace(req.NamaLengkap)
	req.LabelKataKunci = strings.TrimSpace(req.LabelKataKunci)
	req.KataKunci = strings.TrimSpace(req.KataKunci)
	req.DeviceToken = strings.TrimSpace(req.DeviceToken)

	if req.Username == "" || req.Password == "" || req.NamaLengkap == "" || req.NamaSekolah == "" ||
		req.LabelKataKunci == "" || req.KataKunci == "" {
		http.Error(w, `{"error": "Semua field wajib diisi"}`, http.StatusBadRequest)
		return
	}
	if req.DeviceToken == "" {
		http.Error(w, `{"error": "Token perangkat wajib (gunakan halaman daftar dari HP yang sama dengan login)"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengenkripsi password"}`, http.StatusInternalServerError)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, `{"error": "Gagal memulai transaksi database"}`, http.StatusInternalServerError)
		return
	}

	result, err := tx.Exec("INSERT INTO users (username, password, role, is_active) VALUES (?, ?, 'siswa', 0)", req.Username, hashedPassword)
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Username mungkin sudah dipakai"}`, http.StatusConflict)
		return
	}

	userID, _ := result.LastInsertId()

	_, err = tx.Exec(
		"INSERT INTO siswa (user_id, nama_sekolah, nama_lengkap, label_kata_kunci, kata_kunci) VALUES (?, ?, ?, ?, ?)",
		userID, req.NamaSekolah, req.NamaLengkap, req.LabelKataKunci, req.KataKunci,
	)
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Gagal menyimpan data profil siswa"}`, http.StatusInternalServerError)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	_, err = tx.Exec(
		"INSERT INTO user_devices (user_id, device_cookie_token, user_agent, status) VALUES (?, ?, ?, 'pending')",
		userID, req.DeviceToken, userAgent,
	)
	if err != nil {
		tx.Rollback()
		http.Error(w, `{"error": "Gagal menyimpan data perangkat"}`, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Gagal menyelesaikan pendaftaran"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Pendaftaran berhasil dikirim. Tunggu persetujuan admin sebelum bisa login."}`))
}

// GetPendingSiswaRegistrations — admin: siswa dengan is_active = 0
func GetPendingSiswaRegistrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Hanya menerima method GET", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT u.id, u.username, s.nama_lengkap, s.nama_sekolah, s.label_kata_kunci,
		       IFNULL(DATE_FORMAT(u.created_at, '%Y-%m-%d %H:%i'), '') AS created_at
		FROM users u
		JOIN siswa s ON s.user_id = u.id
		WHERE u.role = 'siswa' AND u.is_active = 0
		ORDER BY u.created_at ASC
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil data pendaftar"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []models.PendingSiswaRegistration
	for rows.Next() {
		var p models.PendingSiswaRegistration
		if err := rows.Scan(&p.UserID, &p.Username, &p.NamaLengkap, &p.NamaSekolah, &p.LabelKataKunci, &p.CreatedAt); err != nil {
			http.Error(w, `{"error": "Gagal membaca data"}`, http.StatusInternalServerError)
			return
		}
		list = append(list, p)
	}
	if list == nil {
		list = []models.PendingSiswaRegistration{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// ApproveSiswaRegistration — aktifkan akun + setujui perangkat pending milik siswa ini
func ApproveSiswaRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Hanya menerima method POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.UserID <= 0 {
		http.Error(w, `{"error": "user_id tidak valid"}`, http.StatusBadRequest)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, `{"error": "Gagal memulai transaksi"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var role string
	var active int
	err = tx.QueryRow("SELECT role, is_active FROM users WHERE id = ?", input.UserID).Scan(&role, &active)
	if err != nil {
		http.Error(w, `{"error": "User tidak ditemukan"}`, http.StatusNotFound)
		return
	}
	if role != "siswa" {
		http.Error(w, `{"error": "Hanya akun siswa yang dapat disetujui lewat endpoint ini"}`, http.StatusBadRequest)
		return
	}
	if active != 0 {
		http.Error(w, `{"error": "Akun ini sudah aktif"}`, http.StatusConflict)
		return
	}

	if _, err := tx.Exec("UPDATE users SET is_active = 1 WHERE id = ?", input.UserID); err != nil {
		http.Error(w, `{"error": "Gagal mengaktifkan akun"}`, http.StatusInternalServerError)
		return
	}
	if _, err := tx.Exec("UPDATE user_devices SET status = 'approved' WHERE user_id = ? AND status = 'pending'", input.UserID); err != nil {
		http.Error(w, `{"error": "Gagal menyetujui perangkat"}`, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Gagal menyimpan"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Pendaftar disetujui. Siswa dapat login di perangkat yang didaftarkan."}`))
}

// RejectSiswaRegistration — hapus permanen pendaftar yang belum disetujui (is_active = 0)
func RejectSiswaRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Hanya menerima method POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.UserID <= 0 {
		http.Error(w, `{"error": "user_id tidak valid"}`, http.StatusBadRequest)
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, `{"error": "Gagal memulai transaksi"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var role string
	var active int
	err = tx.QueryRow("SELECT role, is_active FROM users WHERE id = ?", input.UserID).Scan(&role, &active)
	if err != nil {
		http.Error(w, `{"error": "User tidak ditemukan"}`, http.StatusNotFound)
		return
	}
	if role != "siswa" || active != 0 {
		http.Error(w, `{"error": "Hanya pendaftar siswa yang belum disetujui yang dapat ditolak"}`, http.StatusBadRequest)
		return
	}

	tx.Exec("DELETE FROM user_devices WHERE user_id = ?", input.UserID)

	var siswaID int
	if err := tx.QueryRow("SELECT id FROM siswa WHERE user_id = ?", input.UserID).Scan(&siswaID); err == nil && siswaID > 0 {
		tx.Exec("DELETE FROM log_kehadiran WHERE siswa_id = ?", siswaID)
		tx.Exec("DELETE FROM siswa_kelas WHERE siswa_id = ?", siswaID)
		tx.Exec("DELETE FROM siswa WHERE id = ?", siswaID)
	}
	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", input.UserID); err != nil {
		http.Error(w, `{"error": "Gagal menghapus pendaftar"}`, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Gagal menyelesaikan penolakan"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Pendaftaran ditolak dan data dihapus."}`))
}

// GetAllSiswa godoc
// @Summary Tampilkan Semua Data Siswa
// @Description Mengambil daftar seluruh data master siswa
// @Tags 5. Admin - Master Data
// @Produce json
// @Security BearerAuth
// @Success 200 {array} controllers.SiswaData "Data Siswa"
// @Router /admin/siswa/all [get]
// Fitur READ
// Fitur READ
func GetAllSiswa(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Hanya menerima method GET", http.StatusMethodNotAllowed)
        return
    }

    query := `
        SELECT s.id, u.username, s.nama_sekolah, s.nama_lengkap, s.label_kata_kunci
        FROM siswa s
        JOIN users u ON s.user_id = u.id
    `
    
    rows, err := config.DB.Query(query)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengambil data siswa"}`, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var listSiswa []models.SiswaResponse

    for rows.Next() {
        var s models.SiswaResponse
        // PERUBAHAN: Scan diarahkan ke &s.NamaSekolah
        if err := rows.Scan(&s.ID, &s.Username, &s.NamaSekolah, &s.NamaLengkap, &s.LabelKataKunci); err != nil {
            http.Error(w, `{"error": "Gagal membaca baris data"}`, http.StatusInternalServerError)
            return
        }
        listSiswa = append(listSiswa, s)
    }

    if listSiswa == nil {
        listSiswa = []models.SiswaResponse{}
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(listSiswa)
}

// Fitur UPDATE (Ubah Data Siswa)
// Fitur UPDATE (Ubah Data Siswa)
func UpdateSiswa(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "Hanya menerima method PUT", http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, `{"error": "ID User tidak ditemukan di URL"}`, http.StatusBadRequest)
        return
    }

    var req models.SiswaRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
        return
    }

    tx, err := config.DB.Begin()
    if err != nil {
        http.Error(w, `{"error": "Gagal memulai transaksi"}`, http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    _, err = tx.Exec("UPDATE users SET username = ? WHERE id = ?", req.Username, userID)
    if err != nil {
        http.Error(w, `{"error": "Username mungkin sudah dipakai user lain"}`, http.StatusConflict)
        return
    }

    query := `UPDATE siswa SET nama_sekolah = ?, nama_lengkap = ? WHERE user_id = ?`
    _, err = tx.Exec(query, req.NamaSekolah, req.NamaLengkap, userID)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengupdate profil siswa"}`, http.StatusInternalServerError)
        return
    }

    tx.Commit()

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf(`{"message": "Data siswa berhasil diperbarui menjadi %s!"}`, req.NamaLengkap)))
}

// Fitur DELETE (Hapus Data Siswa)
func DeleteSiswa(w http.ResponseWriter, r *http.Request) {
	// ... (kode awal validasi method dan ambil userID tetap sama) ...
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, `{"error": "ID User tidak ditemukan"}`, http.StatusBadRequest)
		return
	}

	// SOFT DELETE: Kita ubah statusnya jadi 0 (Nonaktif/Alumni/Pindah)
	_, err := config.DB.Exec("UPDATE users SET is_active = 0 WHERE id = ?", userID)
	if err != nil {
		http.Error(w, `{"error": "Gagal menonaktifkan siswa."}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Akun Siswa berhasil dinonaktifkan! History absennya tetap aman."}`))
}
