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

	// 1. Enkripsi password siswa
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengenkripsi password"}`, http.StatusInternalServerError)
		return
	}

	// 2. MULAI DATABASE TRANSACTION
	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, `{"error": "Gagal memulai transaksi database"}`, http.StatusInternalServerError)
		return
	}

	// 3. Masukkan ke tabel `users` dulu dengan role 'siswa'
	// Kita gunakan tx.Exec, bukan config.DB.Exec
	result, err := tx.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, 'siswa')", req.Username, hashedPassword)
	if err != nil {
		tx.Rollback() // Batalkan jika gagal
		http.Error(w, `{"error": "Username mungkin sudah dipakai"}`, http.StatusConflict)
		return
	}

	// 4. Ambil ID user yang baru saja dibuat
	userID, _ := result.LastInsertId()

	// 5. Masukkan ke tabel `siswa` menggunakan userID tersebut
	_, err = tx.Exec("INSERT INTO siswa (user_id, nis, nama_lengkap, label_kata_kunci, kata_kunci) VALUES (?, ?, ?, ?, ?)", 
		userID, req.NIS, req.NamaLengkap, req.LabelKataKunci, req.KataKunci)
	
	if err != nil {
		tx.Rollback() // Batalkan jika gagal (data di tabel users otomatis terhapus juga!)
		http.Error(w, `{"error": "Gagal menyimpan data profil siswa. NIS mungkin sudah dipakai."}`, http.StatusInternalServerError)
		return
	}

	// 6. Jika semua lancar, SIMPAN PERMANEN! (Commit)
	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Status 201: Created
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
func GetAllSiswa(w http.ResponseWriter, r *http.Request) {
	// Pastikan hanya menerima request tipe GET
	if r.Method != http.MethodGet {
		http.Error(w, "Hanya menerima method GET", http.StatusMethodNotAllowed)
		return
	}

	// 1. Lakukan query JOIN antara tabel siswa dan users
	query := `
		SELECT s.id, u.username, s.nis, s.nama_lengkap, s.label_kata_kunci
		FROM siswa s
		JOIN users u ON s.user_id = u.id
	`
	
	// config.DB.Query digunakan untuk mengambil BANYAK baris data
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil data siswa"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Pastikan selalu menutup koneksi query setelah selesai

	// 2. Siapkan wadah (array/slice) kosong untuk menampung daftar siswa
	var listSiswa []models.SiswaResponse

	// 3. Looping (putar) untuk membaca data baris demi baris dari database
	for rows.Next() {
		var s models.SiswaResponse
		// Masukkan data dari database ke dalam cetakan 's'
		if err := rows.Scan(&s.ID, &s.Username, &s.NIS, &s.NamaLengkap, &s.LabelKataKunci); err != nil {
			http.Error(w, `{"error": "Gagal membaca baris data"}`, http.StatusInternalServerError)
			return
		}
		// Tambahkan data 's' ke dalam wadah 'listSiswa'
		listSiswa = append(listSiswa, s)
	}

	// 4. Jika datanya kosong, pastikan listSiswa tidak bernilai 'null' saat jadi JSON
	if listSiswa == nil {
		listSiswa = []models.SiswaResponse{}
	}

	// 5. Kirim data kembali ke Admin!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listSiswa)
}

// Fitur UPDATE (Ubah Data Siswa)
func UpdateSiswa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Hanya menerima method PUT", http.StatusMethodNotAllowed)
		return
	}

	// Menangkap ID dari URL (contoh: ?id=1)
	idSiswa := r.URL.Query().Get("id")
	if idSiswa == "" {
		http.Error(w, `{"error": "ID Siswa tidak ditemukan di URL"}`, http.StatusBadRequest)
		return
	}

	var req models.SiswaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
		return
	}

	// Kita hanya mengupdate data profilnya. 
	// (Mengupdate password biasanya dipisah di fitur "Reset Password" khusus)
	query := `UPDATE siswa SET nis = ?, nama_lengkap = ?, label_kata_kunci = ?, kata_kunci = ? WHERE id = ?`
	_, err := config.DB.Exec(query, req.NIS, req.NamaLengkap, req.LabelKataKunci, req.KataKunci, idSiswa)
	
	if err != nil {
		http.Error(w, `{"error": "Gagal mengupdate data profil siswa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Data siswa berhasil diperbarui menjadi %s!"}`, req.NamaLengkap)))
}

// Fitur DELETE (Hapus Data Siswa)
func DeleteSiswa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Hanya menerima method DELETE", http.StatusMethodNotAllowed)
		return
	}

	idSiswa := r.URL.Query().Get("id")
	if idSiswa == "" {
		http.Error(w, `{"error": "ID Siswa tidak ditemukan di URL"}`, http.StatusBadRequest)
		return
	}

	// 1. Cari dulu user_id (induk akun) dari siswa ini
	var userID int
	err := config.DB.QueryRow("SELECT user_id FROM siswa WHERE id = ?", idSiswa).Scan(&userID)
	if err != nil {
		http.Error(w, `{"error": "Data siswa tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	// 2. KEAJAIBAN CASCADE: Kita HANYA menghapus akun induknya (tabel users).
	// Karena di ERD kita sudah menyetel ON DELETE CASCADE, MariaDB akan 
	// secara OTOMATIS menghapus data profilnya di tabel 'siswa' juga!
	_, err = config.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Error(w, `{"error": "Gagal menghapus data siswa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Akun dan Profil siswa berhasil dihapus secara permanen!"}`))
}