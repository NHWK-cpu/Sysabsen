package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"backend-absensi/config"
	"backend-absensi/helpers"
	"backend-absensi/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

// ==========================================
// MANAJEMEN GURU (ADMIN)
// ==========================================

func CreateGuru(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateGuruRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Format data tidak valid"}`, http.StatusBadRequest)
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
	defer tx.Rollback()

	queryUser := `INSERT INTO users (username, password, role) VALUES (?, ?, 'guru')`
	res, err := tx.Exec(queryUser, req.Username, hashedPassword)
	if err != nil {
		http.Error(w, `{"error": "Username mungkin sudah terdaftar. Silakan gunakan yang lain."}`, http.StatusConflict)
		return
	}

	userID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error": "Gagal mendapatkan ID User"}`, http.StatusInternalServerError)
		return
	}

	queryGuru := `INSERT INTO guru (user_id, nama_lengkap, nip, email) VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(queryGuru, userID, req.NamaLengkap, req.NIP, req.Email)
	if err != nil {
		http.Error(w, `{"error": "Gagal menyimpan data profil guru. Pastikan NIP/Data tidak duplikat."}`, http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, `{"error": "Gagal mengonfirmasi data ke database"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Data Guru berhasil ditambahkan!"}`))
}

func DeleteGuru(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, `{"error": "ID User tidak ditemukan"}`, http.StatusBadRequest)
		return
	}

	_, err := config.DB.Exec("UPDATE users SET is_active = 0 WHERE id = ?", userID)
	if err != nil {
		http.Error(w, `{"error": "Gagal menonaktifkan guru."}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Akun Guru berhasil dinonaktifkan! Jadwal dan sejarahnya tetap aman."}`))
}

type GuruRequest struct {
	Username    string `json:"username"`
	NamaLengkap string `json:"nama_lengkap"`
	NIP         string `json:"nip"`
	Email       string `json:"email"`
}

func UpdateGuru(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Hanya menerima method PUT", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, `{"error": "ID User tidak ditemukan di URL"}`, http.StatusBadRequest)
		return
	}

	var req GuruRequest
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

	query := `UPDATE guru SET nip = ?, nama_lengkap = ?, email = ? WHERE user_id = ?`
	_, err = tx.Exec(query, req.NIP, req.NamaLengkap, req.Email, userID)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengupdate profil guru"}`, http.StatusInternalServerError)
		return
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Data guru berhasil diperbarui menjadi %s!"}`, req.NamaLengkap)))
}

// ==========================================
// REFERENSI DATA (UNTUK DROPDOWN GURU)
// ==========================================

func GetMapelForGuru(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, nama_mapel FROM mata_pelajaran WHERE is_active = 1")
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil data mapel"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MapelRes struct {
		ID        int    `json:"id"`
		NamaMapel string `json:"nama_mapel"`
	}
	var list []MapelRes
	for rows.Next() {
		var m MapelRes
		if err := rows.Scan(&m.ID, &m.NamaMapel); err == nil {
			list = append(list, m)
		}
	}
	if list == nil {
		list = []MapelRes{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func GetKelasForGuru(w http.ResponseWriter, r *http.Request) {
	query := `
	SELECT k.id, k.nama_kelas, p.tahun_ajaran, p.semester
	FROM kelas k JOIN periode_belajar p ON k.periode_id = p.id
	WHERE p.status_aktif = 1 ORDER BY k.nama_kelas ASC
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil data kelas"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type KelasRes struct {
		ID        int    `json:"id"`
		NamaKelas string `json:"nama_kelas"`
		TahunAjar string `json:"tahun_ajaran"`
		Semester  string `json:"semester"`
	}
	var list []KelasRes
	for rows.Next() {
		var k KelasRes
		if err := rows.Scan(&k.ID, &k.NamaKelas, &k.TahunAjar, &k.Semester); err == nil {
			list = append(list, k)
		}
	}
	if list == nil {
		list = []KelasRes{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// ==========================================
// OPERASIONAL GURU (ABSENSI & JADWAL)
// ==========================================
func InitOrGetSesi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // WAJIB JSON
	var req struct {
		KelasID int    `json:"kelas_id"`
		MapelID int    `json:"mapel_id"`
		Tanggal string `json:"tanggal"` // Format: YYYY-MM-DD
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Format request salah"}`))
		return
	}

	// VALIDASI: Cegah Svelte mengirim ID 0 (Dropdown kosong)
	if req.KelasID == 0 || req.MapelID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Kelas dan Mata Pelajaran wajib dipilih!"}`))
		return
	}

	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userID := int(userInfo["user_id"].(float64))

	var guruID int
	err := config.DB.QueryRow("SELECT id FROM guru WHERE user_id = ?", userID).Scan(&guruID)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Profil guru tidak ditemukan! Pastikan Anda login sebagai guru."}`))
		return
	}

	var sesiID int
	err = config.DB.QueryRow("SELECT id FROM sesi_pembelajaran WHERE kelas_id = ? AND mapel_id = ? AND DATE(waktu_mulai) = ?", req.KelasID, req.MapelID, req.Tanggal).Scan(&sesiID)

	if err != nil {
		// PERBAIKAN: Hitung waktu mulai dan waktu selesai (otomatis +2 jam)
		currentTime := time.Now()
		waktuMulai := req.Tanggal + currentTime.Format(" 15:04:05")
		waktuSelesai := req.Tanggal + currentTime.Add(2 * time.Hour).Format(" 15:04:05")

		// PERBAIKAN: Masukkan waktu_selesai ke dalam query INSERT
		queryInsert := "INSERT INTO sesi_pembelajaran (kelas_id, mapel_id, guru_id, waktu_mulai, waktu_selesai) VALUES (?, ?, ?, ?, ?)"
		res, errInsert := config.DB.Exec(queryInsert, req.KelasID, req.MapelID, guruID, waktuMulai, waktuSelesai)

		if errInsert != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "Gagal Insert DB: %s"}`, errInsert.Error())))
			return
		}
		newID, _ := res.LastInsertId()
		sesiID = int(newID)
	}

	json.NewEncoder(w).Encode(map[string]int{"sesi_id": sesiID})
}

func GetGuruStats(w http.ResponseWriter, r *http.Request) {
	sesiID := r.URL.Query().Get("sesi_id")
	if sesiID == "" {
		http.Error(w, "sesi_id diperlukan", http.StatusBadRequest)
		return
	}

	var stats struct {
		TotalStudents  int     `json:"total_students"`
		PresentToday   int     `json:"present_today"`
		AbsentToday    int     `json:"absent_today"`
		AttendanceRate float64 `json:"attendance_rate"`
	}

	var kelasID int
	err := config.DB.QueryRow("SELECT kelas_id FROM sesi_pembelajaran WHERE id = ?", sesiID).Scan(&kelasID)
	if err != nil {
		http.Error(w, "Sesi pembelajaran tidak ditemukan", http.StatusNotFound)
		return
	}

	config.DB.QueryRow("SELECT COUNT(*) FROM siswa_kelas WHERE kelas_id = ?", kelasID).Scan(&stats.TotalStudents)
	config.DB.QueryRow("SELECT COUNT(*) FROM log_kehadiran WHERE sesi_id = ? AND status_kehadiran = 'hadir'", sesiID).Scan(&stats.PresentToday)

	stats.AbsentToday = stats.TotalStudents - stats.PresentToday
	if stats.TotalStudents > 0 {
		stats.AttendanceRate = (float64(stats.PresentToday) / float64(stats.TotalStudents)) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func GetSesiSiswaStatus(w http.ResponseWriter, r *http.Request) {
    sesiID := r.URL.Query().Get("sesi_id")
    tanggal := r.URL.Query().Get("tanggal")

    if sesiID == "" || tanggal == "" {
        http.Error(w, `{"error": "sesi_id dan tanggal wajib diisi"}`, http.StatusBadRequest)
        return
    }

    query := `
    SELECT s.id, s.nama_sekolah, s.nama_lengkap, IFNULL(l.status_kehadiran, 'belum_absen') as status, IFNULL(l.waktu_absen, '-') as waktu
    FROM sesi_pembelajaran sp
    JOIN siswa_kelas sk ON sp.kelas_id = sk.kelas_id
    JOIN siswa s ON sk.siswa_id = s.id
    LEFT JOIN log_kehadiran l ON s.id = l.siswa_id AND l.sesi_id = sp.id AND l.tanggal = ?
    WHERE sp.id = ? ORDER BY s.nama_lengkap ASC
    `
    rows, err := config.DB.Query(query, tanggal, sesiID)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengambil data kehadiran"}`, 500)
        return
    }
    defer rows.Close()

    type StudentStatus struct {
        ID          int    `json:"id"`
        NamaSekolah string `json:"nama_sekolah"` // <-- PERUBAHAN
        Nama        string `json:"nama"`
        Status      string `json:"status"`
        Waktu       string `json:"waktu_absen"`
    }

    var list []StudentStatus
    for rows.Next() {
        var s StudentStatus
        if err := rows.Scan(&s.ID, &s.NamaSekolah, &s.Nama, &s.Status, &s.Waktu); err == nil {
            list = append(list, s)
        }
    }
    if list == nil { list = []StudentStatus{} }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(list)
}

func GetJadwalGuru(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method ditolak", http.StatusMethodNotAllowed)
		return
	}

	// KITA BEBASKAN QUERY INI DARI FILTER GURU AGAR FLEKSIBEL
	query := `
	SELECT sp.id, k.nama_kelas, mp.nama_mapel, sp.waktu_mulai, sp.waktu_selesai
	FROM sesi_pembelajaran sp
	JOIN kelas k ON sp.kelas_id = k.id
	JOIN mata_pelajaran mp ON sp.mapel_id = mp.id
	ORDER BY sp.waktu_mulai ASC
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil jadwal mengajar"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listSesi []models.SesiResponse
	for rows.Next() {
		var s models.SesiResponse
		if err := rows.Scan(&s.ID, &s.Kelas, &s.MataPelajaran, &s.WaktuMulai, &s.WaktuSelesai); err == nil {
			listSesi = append(listSesi, s)
		}
	}
	if listSesi == nil { listSesi = []models.SesiResponse{} }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listSesi)
}

func GenerateQRSesi(w http.ResponseWriter, r *http.Request) {
	sesiID := r.URL.Query().Get("sesi_id")
	if sesiID == "" {
		http.Error(w, `{"error": "ID Sesi tidak ditemukan"}`, http.StatusBadRequest)
		return
	}

	var validasiSesi int
	err := config.DB.QueryRow("SELECT id FROM sesi_pembelajaran WHERE id = ?", sesiID).Scan(&validasiSesi)
	if err != nil {
		http.Error(w, `{"error": "Sesi tidak ditemukan!"}`, http.StatusNotFound)
		return
	}

	claims := jwt.MapClaims{
		"sesi_id": sesiID,
		"tipe":    "qr_absen",
		"exp":     time.Now().Add(time.Minute * 5).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWT_KEY)
	if err != nil {
		http.Error(w, `{"error": "Gagal membuat QR Code"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "QR Code berhasil dibuat!", "qr_token": "%s"}`, tokenString)))
}

// ==========================================
// EXPORT & BACKUP
// ==========================================

// ExportDataAbsensi godoc
// Ekspor Data ke Excel dengan format Cross-Tab (1 Sheet = 1 Bulan + Mapel, Tanggal menyamping)
func ExportDataAbsensi(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil data, tambahkan JOIN ke mapel dan sesi untuk mendapatkan nama mapel
	query := `
	SELECT
	s.nama_lengkap,
	mp.nama_mapel,
	LOWER(l.status_kehadiran) as status_kehadiran,
	DATE_FORMAT(l.waktu_absen, '%b %Y') AS bulan_tahun,
	DATE_FORMAT(l.waktu_absen, '%d') AS tanggal
	FROM log_kehadiran l
	JOIN siswa s ON l.siswa_id = s.id
	JOIN sesi_pembelajaran sp ON l.sesi_id = sp.id
	JOIN mata_pelajaran mp ON sp.mapel_id = mp.id
	WHERE l.waktu_absen IS NOT NULL
	ORDER BY l.waktu_absen ASC, s.nama_lengkap ASC
	`
	rows, err := config.DB.Query(query)

	if err != nil {
		http.Error(w, "Gagal mengambil data dari DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 2. Struktur Data Map: map[SheetName] -> map[Nama Siswa] -> map[Tanggal] -> Status
	sheetData := make(map[string]map[string]map[string]string)

	for rows.Next() {
		var namaLengkap, namaMapel, status, bulanTahun, tanggal string
		if err := rows.Scan(&namaLengkap, &namaMapel, &status, &bulanTahun, &tanggal); err == nil {

			// Format Nama Sheet: "Apr 2026 - Matematika"
			sheetName := fmt.Sprintf("%s - %s", bulanTahun, namaMapel)

			// PROTEKSI EXCEL: Potong karakter jika lebih dari 31 huruf
			if len(sheetName) > 31 {
				sheetName = sheetName[:31]
			}

			// Inisialisasi Map jika belum ada
			if sheetData[sheetName] == nil {
				sheetData[sheetName] = make(map[string]map[string]string)
			}
			if sheetData[sheetName][namaLengkap] == nil {
				sheetData[sheetName][namaLengkap] = make(map[string]string)
			}

			// Simpan status absen siswa di tanggal tersebut
			sheetData[sheetName][namaLengkap][tanggal] = status
		}
	}

	// 3. Mulai Buat File Excel
	f := excelize.NewFile()
	firstSheet := true

	if len(sheetData) == 0 {
		f.SetCellValue("Sheet1", "A1", "Belum ada data absensi")
	}

	// 4. Looping untuk membuat Sheet per Kombinasi Bulan & Mapel
	for sheetName, studentMap := range sheetData {
		if firstSheet {
			f.SetSheetName("Sheet1", sheetName)
			firstSheet = false
		} else {
			f.NewSheet(sheetName)
		}

		// --- BIKIN HEADER KOLOM ---
		f.SetCellValue(sheetName, "A1", "Nama Siswa")

		for day := 1; day <= 31; day++ {
			colName, _ := excelize.ColumnNumberToName(day + 1)
			f.SetCellValue(sheetName, fmt.Sprintf("%s1", colName), fmt.Sprintf("%d", day))
		}

		f.SetCellValue(sheetName, "AG1", "Total H")
		f.SetCellValue(sheetName, "AH1", "Total A")
		f.SetCellValue(sheetName, "AI1", "Total I")
		f.SetCellValue(sheetName, "AJ1", "Total S")

		// --- ISI DATA BARIS PER SISWA ---
		rowIndex := 2
		for studentName, daysData := range studentMap {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), studentName)

			countH, countA, countI, countS := 0, 0, 0, 0

			for day := 1; day <= 31; day++ {
				dayStr := fmt.Sprintf("%02d", day)
				status := daysData[dayStr]

				shortStatus := ""
				switch status {
					case "hadir":
						shortStatus = "H"
						countH++
					case "alpa":
						shortStatus = "A"
						countA++
					case "izin":
						shortStatus = "I"
						countI++
					case "sakit":
						shortStatus = "S"
						countS++
				}

				if shortStatus != "" {
					colName, _ := excelize.ColumnNumberToName(day + 1)
					f.SetCellValue(sheetName, fmt.Sprintf("%s%d", colName, rowIndex), shortStatus)
				}
			}

			f.SetCellValue(sheetName, fmt.Sprintf("AG%d", rowIndex), countH)
			f.SetCellValue(sheetName, fmt.Sprintf("AH%d", rowIndex), countA)
			f.SetCellValue(sheetName, fmt.Sprintf("AI%d", rowIndex), countI)
			f.SetCellValue(sheetName, fmt.Sprintf("AJ%d", rowIndex), countS)

			rowIndex++
		}
	}

	// 5. Simpan dan Export
	fileName := "Laporan_Absen_" + time.Now().Format("2006-01-02") + ".xlsx"
	err = f.SaveAs(fileName)
	if err != nil {
		http.Error(w, "Gagal membuat file Excel lokal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	srv, errDrive := helpers.InitDriveService()
	if errDrive == nil {
		folderID := "1akOpzgXyB7r8oO7JSolBnfIGwL46xV__"
		_ = helpers.UploadToDrive(srv, fileName, folderID)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	f.Write(w)

	time.Sleep(500 * time.Millisecond)
	os.Remove(fileName)
}

