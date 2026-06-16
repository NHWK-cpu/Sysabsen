package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5" // Tambahkan ini untuk membaca context
	"net/http"
	"os"
	"os/exec"
	"strconv" // Tambahkan ini jika belum ada
	"strings"
	"time"

	"backend-absensi/config"
	"backend-absensi/helpers"
	"backend-absensi/models" // Import package models yang baru
)

// =========================================
// RESTORE DATABASE
// =========================================

func BackupDatabase(w http.ResponseWriter, r *http.Request) {
	// 1. Kasih timeout 15 detik untuk proses ekspor
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fileName := fmt.Sprintf("Backup_Full_%s.sql", time.Now().Format("2006-01-02_15-04-05"))

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root" // Fallback aman jika DB_USER di .env kosong
	}
	dbName := os.Getenv("DB_NAME")

	// Karena di VPS, mysqldump bisa langsung jalan tanpa -h
	cmd := exec.CommandContext(ctx, "mysqldump", "-u", dbUser, dbName, "--result-file="+fileName)

	// Inject password lewat env agar aman dari interaktif prompt
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+os.Getenv("DB_PASSWORD"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Gagal: Timeout saat membuat file backup di VPS.", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, fmt.Sprintf("Gagal dump database:\nError: %s\nOutput: %s", err.Error(), string(output)), http.StatusInternalServerError)
		return
	}

	// 2. TANGKAP ERROR GOOGLE DRIVE (Ini yang bikin Panic)
	srv, errDriveInit := helpers.InitDriveService()
	if errDriveInit != nil {
		os.Remove(fileName) // Hapus file lokal agar VPS tidak penuh
		http.Error(w, "Sistem gagal terhubung ke Google Drive. Pastikan file credentials.json ada di VPS: "+errDriveInit.Error(), http.StatusInternalServerError)
		return
	}

	errDriveUpload := helpers.UploadToDrive(srv, fileName, "1OpRprWCk2MgurUclTWdTR43EaoKuOIGH")

	// Hapus file dump setelah upload selesai
	time.Sleep(500 * time.Millisecond)
	os.Remove(fileName)

	if errDriveUpload != nil {
		http.Error(w, "Gagal upload backup ke Google Drive: "+errDriveUpload.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Backup Database Berhasil Diupload ke Google Drive!"))
}

func RestoreDatabase(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("file_id")
	if fileID == "" {
		http.Error(w, "file_id wajib diisi", http.StatusBadRequest)
		return
	}

	localFile := "restore_temp.sql"

	// TANGKAP ERROR INIT DRIVE
	srv, errDriveInit := helpers.InitDriveService()
	if errDriveInit != nil {
		http.Error(w, "Sistem gagal terhubung ke Google Drive: "+errDriveInit.Error(), http.StatusInternalServerError)
		return
	}

	if err := helpers.DownloadFromDrive(srv, fileID, localFile); err != nil {
		http.Error(w, "Gagal download file SQL dari Drive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbName := os.Getenv("DB_NAME")

	cmd := exec.CommandContext(ctx, "mysql", "-u", dbUser, dbName)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+os.Getenv("DB_PASSWORD"))

	sqlFile, _ := os.Open(localFile)
	cmd.Stdin = sqlFile

	output, err := cmd.CombinedOutput()

	sqlFile.Close()
	os.Remove(localFile)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Gagal: Timeout saat melakukan proses Restore di MariaDB.", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, fmt.Sprintf("Gagal restore database:\nError: %s\nOutput: %s", err.Error(), string(output)), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Database Berhasil Direstore ke Kondisi Backup!"))
}

// ==========================================
// FITUR PERIODE BELAJAR
// ==========================================

func CreatePeriode(w http.ResponseWriter, r *http.Request) {
	var p models.PeriodeBelajar
	json.NewDecoder(r.Body).Decode(&p)

	// Tambahkan semester ke dalam query INSERT
	_, err := config.DB.Exec("INSERT INTO periode_belajar (tahun_ajaran, semester, status_aktif) VALUES (?, ?, ?)", p.TahunAjar, p.Semester, p.StatusAktif)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Periode berhasil dibuat"))
}

func GetAllPeriode(w http.ResponseWriter, r *http.Request) {
	// Tambahkan semester ke dalam query SELECT
	rows, err := config.DB.Query("SELECT id, tahun_ajaran, semester, status_aktif FROM periode_belajar")
	if err != nil {
		http.Error(w, "Gagal query periode: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []models.PeriodeBelajar
	for rows.Next() {
		var p models.PeriodeBelajar
		if err := rows.Scan(&p.ID, &p.TahunAjar, &p.Semester, &p.StatusAktif); err != nil {
			continue
		}
		list = append(list, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func UpdatePeriode(w http.ResponseWriter, r *http.Request) {
	// Karena kita cuma butuh ID dan status_aktif, kita pakai struct anonim
	// atau bisa juga pakai models.PeriodeBelajar kalau mau
	var p struct {
		ID          int `json:"id"`
		StatusAktif int `json:"status_aktif"`
	}

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	// Eksekusi update hanya pada kolom status_aktif
	_, err := config.DB.Exec("UPDATE periode_belajar SET status_aktif = ? WHERE id = ?", p.StatusAktif, p.ID)
	if err != nil {
		http.Error(w, "Gagal mengubah status periode: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Status periode berhasil diperbarui"))
}

func DeletePeriode(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// CEK MANUAL: Apakah ada kelas yang nyantol ke periode ini?
	var count int
	errCheck := config.DB.QueryRow("SELECT COUNT(*) FROM kelas WHERE periode_id = ?", id).Scan(&count)
	if errCheck == nil && count > 0 {
		http.Error(w, "Gagal: Periode ini tidak bisa dihapus karena masih memiliki Kelas di dalamnya.", http.StatusConflict)
		return
	}

	_, err := config.DB.Exec("DELETE FROM periode_belajar WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Gagal hapus periode dari database: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Periode berhasil dihapus"))
}

// ==========================================
// FITUR PIVOT SISWA-KELAS (Assign Siswa ke Kelas)
// ==========================================

// ==========================================
// FITUR PIVOT SISWA-KELAS (Assign Siswa ke Kelas)
// ==========================================

func GetKelasBySiswa(w http.ResponseWriter, r *http.Request) {
	userSiswaID := r.URL.Query().Get("siswa_id")

	// 1. TERJEMAHKAN ID DARI USERS KE SISWA
	var realSiswaID int
	err := config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", userSiswaID).Scan(&realSiswaID)
	if err != nil {
		http.Error(w, "Gagal: Profil siswa tidak ditemukan.", http.StatusNotFound)
		return
	}

	// 2. AMBIL DAFTAR KELAS BESERTA PERIODENYA
	query := `
	SELECT k.id, k.nama_kelas, p.tahun_ajaran, p.semester
	FROM siswa_kelas sk
	JOIN kelas k ON sk.kelas_id = k.id
	JOIN periode_belajar p ON k.periode_id = p.id
	WHERE sk.siswa_id = ?
	`
	rows, err := config.DB.Query(query, realSiswaID)
	if err != nil {
		http.Error(w, "Gagal query daftar kelas: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	type KelasResponse struct {
		ID        int    `json:"id"`
		NamaKelas string `json:"name"`
		TahunAjar string `json:"tahun_ajaran"`
		Semester  string `json:"semester"`
	}

	var list []KelasResponse
	for rows.Next() {
		var k KelasResponse
		if err := rows.Scan(&k.ID, &k.NamaKelas, &k.TahunAjar, &k.Semester); err == nil {
			list = append(list, k)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func GetSiswaByKelas(w http.ResponseWriter, r *http.Request) {
	kelasID := r.URL.Query().Get("kelas_id")

	query := `
    SELECT u.id as user_id, s.nama_lengkap, s.nama_sekolah
    FROM siswa_kelas sk
    JOIN siswa s ON sk.siswa_id = s.id
    JOIN users u ON s.user_id = u.id
    WHERE sk.kelas_id = ?
    ORDER BY s.nama_lengkap ASC
    `
	rows, err := config.DB.Query(query, kelasID)
	if err != nil {
		http.Error(w, "Gagal mengambil data siswa di kelas: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	type SiswaKelasResponse struct {
		UserID      int    `json:"user_id"`
		NamaLengkap string `json:"nama_lengkap"`
		NamaSekolah string `json:"nama_sekolah"`
	}

	var list []SiswaKelasResponse
	for rows.Next() {
		var s SiswaKelasResponse
		// PERUBAHAN: Scan ke &s.NamaSekolah
		if err := rows.Scan(&s.UserID, &s.NamaLengkap, &s.NamaSekolah); err == nil {
			list = append(list, s)
		}
	}

	if list == nil {
		list = []SiswaKelasResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func AssignSiswaToKelas(w http.ResponseWriter, r *http.Request) {
	var sk struct {
		SiswaUserID int `json:"siswa_id"`
		KelasID     int `json:"kelas_id"`
	}
	json.NewDecoder(r.Body).Decode(&sk)

	// AMBIL DATA USER YANG SEDANG LOGIN DARI TOKEN
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	role := userInfo["role"].(string)
	loggedInUserID := int(userInfo["user_id"].(float64))

	// CEK OTORISASI JIKA YANG LOGIN ADALAH GURU
	if role == "guru" {
		var waliGuruID int
		// Cari ID guru berdasarkan user_id, lalu cocokkan dengan wali_guru_id di tabel kelas
		errCheck := config.DB.QueryRow(`
            SELECT k.wali_guru_id FROM kelas k 
            JOIN guru g ON k.wali_guru_id = g.id 
            WHERE k.id = ? AND g.user_id = ?`, sk.KelasID, loggedInUserID).Scan(&waliGuruID)

		if errCheck != nil {
			http.Error(w, "Akses ditolak: Anda bukan wali dari kelas ini!", http.StatusForbidden)
			return
		}
	}

	// TERJEMAHKAN: Cari siswa.id yang asli berdasarkan users.id
	var realSiswaID int
	err := config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", sk.SiswaUserID).Scan(&realSiswaID)
	if err != nil {
		http.Error(w, "Gagal: Profil siswa tidak ditemukan di tabel master siswa.", http.StatusNotFound)
		return
	}

	// EKSEKUSI
	_, err = config.DB.Exec("INSERT INTO siswa_kelas (siswa_id, kelas_id) VALUES (?, ?)", realSiswaID, sk.KelasID)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
			http.Error(w, "Siswa sudah terdaftar di kelas ini. Pilih siswa lain atau cek kembali daftar siswa pada kelas tersebut.", http.StatusConflict)
			return
		}

		http.Error(w, "Gagal menambahkan siswa ke kelas. Silakan coba lagi.", http.StatusInternalServerError)
		return
	}

	// CATAT LOG OTOMATIS
	deskripsi := fmt.Sprintf("Menambahkan siswa (User ID: %d) ke dalam Kelas ID: %d", sk.SiswaUserID, sk.KelasID)
	go catatLog(loggedInUserID, role, "ASSIGN_SISWA", deskripsi) // Pakai 'go' agar berjalan di background tanpa memblokir respon

	w.Write([]byte("Siswa berhasil dimasukkan ke kelas"))
}

func UpdateSiswaKelas(w http.ResponseWriter, r *http.Request) {
	var sk struct {
		SiswaUserID int `json:"siswa_id"`
		OldKelasID  int `json:"old_kelas_id"`
		NewKelasID  int `json:"new_kelas_id"`
	}
	json.NewDecoder(r.Body).Decode(&sk)

	// AMBIL DATA PELAKU DARI TOKEN
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	role := userInfo["role"].(string)
	loggedInUserID := int(userInfo["user_id"].(float64))

	// CEK OTORISASI KHUSUS GURU
	if role == "guru" {
		// Logikanya: Guru hanya boleh memutasi siswa JIKA dia adalah wali dari kelas asal (OldKelasID)
		var waliGuruID int
		errCheck := config.DB.QueryRow(`
            SELECT k.wali_guru_id FROM kelas k 
            JOIN guru g ON k.wali_guru_id = g.id 
            WHERE k.id = ? AND g.user_id = ?`, sk.OldKelasID, loggedInUserID).Scan(&waliGuruID)

		if errCheck != nil {
			http.Error(w, "Akses ditolak: Anda bukan wali dari kelas asal siswa ini!", http.StatusForbidden)
			return
		}
	}

	// 1. TERJEMAHKAN ID SISWA
	var realSiswaID int
	err := config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", sk.SiswaUserID).Scan(&realSiswaID)
	if err != nil {
		http.Error(w, "Gagal: Profil siswa tidak ditemukan.", http.StatusNotFound)
		return
	}

	// 2. EKSEKUSI UPDATE
	_, err = config.DB.Exec("UPDATE siswa_kelas SET kelas_id = ? WHERE siswa_id = ? AND kelas_id = ?", sk.NewKelasID, realSiswaID, sk.OldKelasID)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062") {
			http.Error(w, "Siswa sudah terdaftar di kelas tujuan. Pilih kelas lain atau cek kembali data kelas siswa tersebut.", http.StatusConflict)
			return
		}

		http.Error(w, "Gagal memperbarui kelas siswa. Silakan coba lagi.", http.StatusInternalServerError)
		return
	}

	// 3. CATAT LOG MUTASI
	deskripsi := fmt.Sprintf("Mutasi siswa (User ID: %d) dari Kelas ID: %d menuju Kelas ID: %d", sk.SiswaUserID, sk.OldKelasID, sk.NewKelasID)
	go catatLog(loggedInUserID, role, "UPDATE_SISWA_KELAS", deskripsi)

	w.Write([]byte("Data kelas siswa berhasil diperbarui (Mutasi Berhasil)"))
}

func RemoveSiswaFromKelas(w http.ResponseWriter, r *http.Request) {
	userSiswaID := r.URL.Query().Get("siswa_id")
	kelasID := r.URL.Query().Get("kelas_id")
	deleteLogs := r.URL.Query().Get("delete_logs") // Menangkap parameter "true" atau "false"

	kelasIDInt, _ := strconv.Atoi(kelasID)

	// AMBIL DATA USER YANG SEDANG LOGIN
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	role := userInfo["role"].(string)
	loggedInUserID := int(userInfo["user_id"].(float64))

	// CEK OTORISASI GURU
	if role == "guru" {
		var waliGuruID int
		errCheck := config.DB.QueryRow(`
            SELECT k.wali_guru_id FROM kelas k 
            JOIN guru g ON k.wali_guru_id = g.id 
            WHERE k.id = ? AND g.user_id = ?`, kelasIDInt, loggedInUserID).Scan(&waliGuruID)

		if errCheck != nil {
			http.Error(w, "Akses ditolak: Anda bukan wali dari kelas ini!", http.StatusForbidden)
			return
		}
	}

	// TERJEMAHKAN ID
	var realSiswaID int
	err := config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", userSiswaID).Scan(&realSiswaID)
	if err != nil {
		http.Error(w, "Gagal: Profil siswa tidak ditemukan.", http.StatusNotFound)
		return
	}

	// MULAI TRANSAKSI
	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, "Gagal memulai pemrosesan data", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Akan dibatalkan jika ada error sebelum tx.Commit()

	// JIKA GURU MEMILIH MENGHAPUS LOG ABSENSI
	if deleteLogs == "true" {
		queryDeleteLogs := `
            DELETE l FROM log_kehadiran l
            JOIN sesi_pembelajaran sp ON l.sesi_id = sp.id
            WHERE l.siswa_id = ? AND sp.kelas_id = ?
        `
		_, err = tx.Exec(queryDeleteLogs, realSiswaID, kelasID)
		if err != nil {
			http.Error(w, "Gagal menghapus riwayat absen: "+err.Error(), 500)
			return
		}
	}

	// HAPUS DARI KELAS
	_, err = tx.Exec("DELETE FROM siswa_kelas WHERE siswa_id = ? AND kelas_id = ?", realSiswaID, kelasID)
	if err != nil {
		http.Error(w, "Gagal mengeluarkan siswa dari kelas: "+err.Error(), 500)
		return
	}

	tx.Commit()

	// CATAT LOG OTOMATIS
	deskripsi := fmt.Sprintf("Mengeluarkan siswa (User ID: %s) dari Kelas ID: %s. Hapus Log: %s", userSiswaID, kelasID, deleteLogs)
	go catatLog(loggedInUserID, role, "REMOVE_SISWA", deskripsi)

	w.Write([]byte("Siswa berhasil dikeluarkan dari kelas tersebut"))
}

// ==========================================
// FITUR MATA PELAJARAN (MAPEL)
// ==========================================

// CreateMapel godoc
// @Summary Tambah Mata Pelajaran Baru
// @Tags 5. Admin - Master Data
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body controllers.CreateMapelReq true "Nama Mapel"
// @Success 200 {object} controllers.SuccessMessage "Mapel berhasil ditambahkan"
// @Router /admin/mapel/create [post]
func CreateMapel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Gunakan method POST", http.StatusMethodNotAllowed)
		return
	}

	var mapel models.Mapel // Menggunakan struct dari package models
	if err := json.NewDecoder(r.Body).Decode(&mapel); err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("INSERT INTO mata_pelajaran (nama_mapel) VALUES (?)", mapel.NamaMapel)
	if err != nil {
		http.Error(w, "Gagal menyimpan ke database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Berhasil! Mata Pelajaran baru telah ditambahkan."))
}

func GetAllMapel(w http.ResponseWriter, r *http.Request) {
	// Tambahkan is_active ke dalam SELECT
	rows, err := config.DB.Query("SELECT id, nama_mapel, IFNULL(is_active, 1) FROM mata_pelajaran")
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Bikin struct anonim aja khusus buat balikan JSON ini
	type MapelRes struct {
		ID        int    `json:"id"`
		NamaMapel string `json:"nama_mapel"`
		IsActive  int    `json:"is_active"`
	}

	var listMapel []MapelRes
	for rows.Next() {
		var m MapelRes
		if err := rows.Scan(&m.ID, &m.NamaMapel, &m.IsActive); err == nil {
			listMapel = append(listMapel, m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listMapel)
}

func ToggleStatusMapel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int `json:"id"`
		IsActive int `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("UPDATE mata_pelajaran SET is_active = ? WHERE id = ?", req.IsActive, req.ID)
	if err != nil {
		http.Error(w, "Gagal mengubah status mapel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Status mapel berhasil diperbarui"))
}

func UpdateMapel(w http.ResponseWriter, r *http.Request) {
	var m models.Mapel
	json.NewDecoder(r.Body).Decode(&m)
	_, err := config.DB.Exec("UPDATE mata_pelajaran SET nama_mapel = ? WHERE id = ?", m.NamaMapel, m.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Mapel berhasil diperbarui"))
}

func DeleteMapel(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// CEK MANUAL: Apakah mapel ini sudah pernah dipakai untuk bikin Sesi Pembelajaran?
	// Catatan: Sesuaikan nama tabel 'sesi_pembelajaran' dan 'mapel_id' dengan ERD asli kamu ya.
	var count int
	errCheck := config.DB.QueryRow("SELECT COUNT(*) FROM sesi_pembelajaran WHERE mapel_id = ?", id).Scan(&count)
	if errCheck == nil && count > 0 {
		http.Error(w, "Gagal: Mapel tidak bisa dihapus karena sudah memiliki riwayat Sesi Pembelajaran / Absensi. Sebaiknya nonaktifkan saja jika tidak dipakai.", http.StatusConflict)
		return
	}

	_, err := config.DB.Exec("DELETE FROM mata_pelajaran WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Gagal hapus mapel dari database: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Mapel berhasil dihapus"))
}

// ==========================================
// FITUR KELAS
// ==========================================

// CreateKelas godoc
// @Summary Tambah Kelas Baru
// @Tags 5. Admin - Master Data
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body controllers.CreateKelasReq true "Nama Kelas"
// @Success 200 {object} controllers.SuccessMessage "Kelas berhasil ditambahkan"
// @Router /admin/kelas/create [post]
func CreateKelas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Gunakan method POST", http.StatusMethodNotAllowed)
		return
	}

	var kelas models.Kelas
	if err := json.NewDecoder(r.Body).Decode(&kelas); err != nil {
		http.Error(w, "Format JSON tidak valid", http.StatusBadRequest)
		return
	}

	// Validasi tambahan agar tidak dikirim kosong
	if kelas.PeriodeID == 0 || kelas.NamaKelas == "" {
		http.Error(w, "periode_id dan nama_kelas wajib diisi", http.StatusBadRequest)
		return
	}

	// Eksekusi ke Database dengan menyertakan periode_id
	_, err := config.DB.Exec("INSERT INTO kelas (periode_id, nama_kelas) VALUES (?, ?)", kelas.PeriodeID, kelas.NamaKelas)
	if err != nil {
		http.Error(w, "Gagal menyimpan ke database (pastikan periode_id valid): "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Berhasil! Kelas baru telah ditambahkan pada periode tersebut."))
}

func GetAllKelas(w http.ResponseWriter, r *http.Request) {
	// 1. Gunakan LEFT JOIN ke tabel guru untuk menarik g.user_id
	query := `
        SELECT k.id, k.periode_id, g.user_id, k.nama_kelas 
        FROM kelas k
        LEFT JOIN guru g ON k.wali_guru_id = g.id
    `
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listKelas []models.Kelas
	for rows.Next() {
		var k models.Kelas
		// 2. Scan HARUS 4 variabel sesuai dengan urutan SELECT di atas!
		// k.WaliGuruID sekarang akan berisi g.user_id, sehingga cocok dengan Frontend Svelte
		if err := rows.Scan(&k.ID, &k.PeriodeID, &k.WaliGuruID, &k.NamaKelas); err != nil {
			fmt.Println("Error scan kelas:", err) // Tambahkan ini agar kalau error kelihatan di terminal VPS/Lokal
			continue
		}
		listKelas = append(listKelas, k)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listKelas)
}

func UpdateKelas(w http.ResponseWriter, r *http.Request) {
	var k models.Kelas
	json.NewDecoder(r.Body).Decode(&k)
	_, err := config.DB.Exec("UPDATE kelas SET periode_id = ?, nama_kelas = ? WHERE id = ?", k.PeriodeID, k.NamaKelas, k.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Kelas berhasil diperbarui"))
}

func DeleteKelas(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// CEK MANUAL: Apakah ada siswa yang masih terdaftar di kelas ini?
	var count int
	errCheck := config.DB.QueryRow("SELECT COUNT(*) FROM siswa_kelas WHERE kelas_id = ?", id).Scan(&count)
	if errCheck == nil && count > 0 {
		http.Error(w, "Gagal: Kelas tidak bisa dihapus karena masih ada siswa di dalamnya. Keluarkan dulu semua siswa dari kelas ini.", http.StatusConflict)
		return
	}

	_, err := config.DB.Exec("DELETE FROM kelas WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Gagal hapus kelas dari database: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Kelas berhasil dihapus"))
}

// GetDashboardStats godoc
// @Summary Statistik Dashboard Admin
// @Description Menampilkan jumlah total akun, akun aktif (90 hari), inaktif, dan antrean device
// @Tags 4. Admin - Dashboard & Device
// @Produce json
// @Security BearerAuth
// @Success 200 {object} controllers.AdminStatsResponse "Data Statistik"
// @Router /admin/dashboard/stats [get]
// GetDashboardStats: API untuk Dashboard Admin (Sesuai Desain UI)
func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	var stats struct {
		TotalUsers                int `json:"total_users"`
		ActiveUsers               int `json:"active_users"`
		InactiveUsers             int `json:"inactive_users"`
		PendingDevices            int `json:"pending_devices"`
		PendingSiswaRegistrations int `json:"pending_siswa_registrations"`
		RecentLogins              []struct {
			Time   string `json:"time"`
			User   string `json:"user"`
			Role   string `json:"role"`
			Status string `json:"status"`
		} `json:"recent_logins"`
	}

	// 1. Hitung Total Semua User (TERMASUK YANG NONAKTIF)
	// Hapus filter is_active agar angka total tidak membingungkan
	config.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)

	// 2. Hitung Active Users (Hanya yang is_active = 1 dan login dalam 24 jam terakhir)
	config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE last_login >= NOW() - INTERVAL 1 DAY AND is_active = 1").Scan(&stats.ActiveUsers)

	// 3. Hitung Inactive Users
	// Secara matematis, user yang is_active = 0 otomatis akan masuk ke hitungan ini
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	// 4. Hitung Pending Device
	config.DB.QueryRow("SELECT COUNT(*) FROM user_devices WHERE status = 'pending'").Scan(&stats.PendingDevices)

	// 4b. Siswa mendaftar mandiri, menunggu persetujuan admin
	config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'siswa' AND is_active = 0").Scan(&stats.PendingSiswaRegistrations)

	// 5. Ambil 5 History Login Terbaru (Hanya yang is_active = 1 yang boleh tampil di "Login Terbaru")
	query := `
	SELECT
	DATE_FORMAT(last_login, '%H:%i WIB') as time,
	username as user,
	role,
	'Online' as status
	FROM users
	WHERE last_login IS NOT NULL AND is_active = 1
	ORDER BY last_login DESC
	LIMIT 5
	`
	rows, err := config.DB.Query(query)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rl struct {
				Time   string `json:"time"`
				User   string `json:"user"`
				Role   string `json:"role"`
				Status string `json:"status"`
			}
			rows.Scan(&rl.Time, &rl.User, &rl.Role, &rl.Status)
			stats.RecentLogins = append(stats.RecentLogins, rl)
		}
	}

	if stats.RecentLogins == nil {
		stats.RecentLogins = []struct {
			Time   string `json:"time"`
			User   string `json:"user"`
			Role   string `json:"role"`
			Status string `json:"status"`
		}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetInactiveUsers: Mengambil user yang tidak login lebih dari 90 hari
func GetInactiveUsers(w http.ResponseWriter, r *http.Request) {
	// Menggunakan COALESCE untuk mengambil nama dari siswa atau guru
	// Menggunakan IFNULL untuk menangani last_login yang masih kosong
	query := `
	SELECT u.id,
	COALESCE(s.nama_lengkap, g.nama_lengkap, 'Belum Ada Nama') as nama_lengkap,
	u.username,
	u.role,
	IFNULL(u.last_login, 'Belum Pernah') as last_login,
	IF(u.last_login IS NULL, 0, DATEDIFF(NOW(), u.last_login)) as days_inactive
	FROM users u
	LEFT JOIN siswa s ON u.id = s.user_id
	LEFT JOIN guru g ON u.id = g.user_id
	WHERE (u.last_login < DATE_SUB(NOW(), INTERVAL 90 DAY) OR u.last_login IS NULL)
	AND u.role != 'admin'
	AND u.is_active = 1` // <-- TAMBAHKAN FILTER INI DI SINI

	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, "Gagal query data inaktif: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var list []models.UserDashboardRow
	for rows.Next() {
		var u models.UserDashboardRow
		// Pastikan urutan Scan sesuai dengan urutan SELECT di atas
		err := rows.Scan(&u.ID, &u.NamaLengkap, &u.Username, &u.Role, &u.LastLogin, &u.DaysInactive)
		if err != nil {
			continue
		}
		list = append(list, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetAllUsers: Mengambil semua user untuk tabel utama
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Kita buat struct khusus untuk endpoint ini agar sesuai tarikan Svelte
	type UserResponse struct {
		ID          int    `json:"id"`
		NamaLengkap string `json:"nama_lengkap"`
		Username    string `json:"username"`
		Role        string `json:"role"`
		Identifier  string `json:"identifier"` // <-- Menampung Nama Sekolah atau NIP
		Email       string `json:"email"`
		LastLogin   string `json:"last_login"`
		IsActive    int    `json:"is_active"`
	}

	// PERBAIKAN QUERY:
	// Tambahkan COALESCE(g.email, '') agar tidak error saat meload data siswa yang tidak punya email
	query := `
    SELECT u.id,
    COALESCE(s.nama_lengkap, g.nama_lengkap, 'User Baru') as nama_lengkap,
    u.username,
    u.role,
    COALESCE(s.nama_sekolah, g.nip, '-') as identifier, 
    COALESCE(g.email, '') as email,
    IFNULL(u.last_login, '-') as last_login,
    u.is_active
    FROM users u
    LEFT JOIN siswa s ON u.id = s.user_id
    LEFT JOIN guru g ON u.id = g.user_id
    WHERE u.role != 'admin'
    ORDER BY u.last_login DESC`

	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, "Gagal query semua user: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	var list []UserResponse
	for rows.Next() {
		var u UserResponse
		// <-- TAMBAHAN: Jangan lupa tangkap &u.Email di urutan ke-6
		err := rows.Scan(&u.ID, &u.NamaLengkap, &u.Username, &u.Role, &u.Identifier, &u.Email, &u.LastLogin, &u.IsActive)
		if err != nil {
			continue
		}
		list = append(list, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func catatLog(userID int, role, action, deskripsi string) {
	config.DB.Exec("INSERT INTO activity_logs (user_id, role, action, deskripsi) VALUES (?, ?, ?, ?)", userID, role, action, deskripsi)
}
