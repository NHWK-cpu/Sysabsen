package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

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

    result, err := tx.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, 'siswa')", req.Username, hashedPassword)
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
