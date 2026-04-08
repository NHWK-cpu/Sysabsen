package controllers

import (
	"fmt"
	"time"
	"encoding/json"
	"net/http"
	"os/exec" 
	"os"

	"backend-absensi/config"
	"backend-absensi/helpers"
	"backend-absensi/models"

	"github.com/xuri/excelize/v2"
	"github.com/golang-jwt/jwt/v5"
)

// GetGuruStats godoc
// @Summary Ringkasan Statistik Kelas
// @Description Menampilkan total murid terdaftar, hadir, dan rasio kehadiran
// @Tags 3. Operasional Guru
// @Produce json
// @Security BearerAuth
// @Param sesi_id query int true "ID Sesi Pembelajaran"
// @Success 200 {object} controllers.GuruStatsResponse "Data Ditemukan"
// @Router /guru/dashboard/stats [get]
func GetGuruStats(w http.ResponseWriter, r *http.Request) {
	// Ambil ID sesi dari query parameter URL (?sesi_id=1)
	sesiID := r.URL.Query().Get("sesi_id")
	if sesiID == "" {
		http.Error(w, "sesi_id diperlukan untuk melihat statistik kelas", http.StatusBadRequest)
		return
	}

	// Struktur JSON yang menyesuaikan persis dengan Card di UI Figma kamu
	var stats struct {
		TotalStudents  int     `json:"total_students"`
		PresentToday   int     `json:"present_today"`
		AbsentToday    int     `json:"absent_today"`
		AttendanceRate float64 `json:"attendance_rate"`
	}

	// 1. Cari kelas_id dari sesi ini untuk mengetahui siapa saja yang terdaftar
	var kelasID int
	err := config.DB.QueryRow("SELECT kelas_id FROM sesi_pembelajaran WHERE id = ?", sesiID).Scan(&kelasID)
	if err != nil {
		http.Error(w, "Sesi pembelajaran tidak ditemukan", http.StatusNotFound)
		return
	}

	// 2. Hitung Total Siswa (Enrolled in class) dari tabel pivot
	err = config.DB.QueryRow("SELECT COUNT(*) FROM siswa_kelas WHERE kelas_id = ?", kelasID).Scan(&stats.TotalStudents)
	if err != nil {
		stats.TotalStudents = 0
	}

	// 3. Hitung Hadir Hari Ini (Present Today)
	err = config.DB.QueryRow("SELECT COUNT(*) FROM log_kehadiran WHERE sesi_id = ? AND status = 'hadir'", sesiID).Scan(&stats.PresentToday)
	if err != nil {
		stats.PresentToday = 0
	}

	// 4. Kalkulasi Absent & Persentase
	// Yang tidak absen (atau status selain hadir) dianggap Absent di UI ini
	stats.AbsentToday = stats.TotalStudents - stats.PresentToday

	// Cegah pembagian dengan nol jika kelas masih kosong
	if stats.TotalStudents > 0 {
		stats.AttendanceRate = (float64(stats.PresentToday) / float64(stats.TotalStudents)) * 100
	} else {
		stats.AttendanceRate = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetSesiSiswaStatus godoc
// @Summary Detail Status Kehadiran Per Siswa
// @Description Menampilkan daftar nama siswa dan status absensinya (Hadir/Belum)
// @Tags 3. Operasional Guru
// @Produce json
// @Security BearerAuth
// @Param sesi_id query int true "ID Sesi Pembelajaran"
// @Success 200 {array} controllers.StudentStatusList "Daftar Kehadiran"
// @Router /guru/dashboard/attendance-list [get]
// GetSesiSiswaStatus: Menampilkan daftar kehadiran seluruh siswa di satu sesi tertentu
func GetSesiSiswaStatus(w http.ResponseWriter, r *http.Request) {
	// Ambil ID sesi dari query parameter (?sesi_id=1)
	sesiID := r.URL.Query().Get("sesi_id")
	if sesiID == "" {
		http.Error(w, `{"error": "sesi_id wajib diisi"}`, http.StatusBadRequest)
		return
	}

	// Query SQL yang menggabungkan:
	// 1. Daftar siswa yang terdaftar di kelas sesi tersebut (siswa_kelas)
	// 2. Data kehadiran mereka di sesi itu (log_kehadiran) jika ada
	query := `
		SELECT 
			s.nis, 
			s.nama_lengkap, 
			IFNULL(l.status_kehadiran, 'belum_absen') as status,
			IFNULL(l.waktu_absen, '-') as waktu
		FROM sesi_pembelajaran sp
		JOIN siswa_kelas sk ON sp.kelas_id = sk.kelas_id
		JOIN siswa s ON sk.siswa_id = s.id
		LEFT JOIN log_kehadiran l ON s.id = l.siswa_id AND l.sesi_id = sp.id
		WHERE sp.id = ?
		ORDER BY s.nama_lengkap ASC
	`

	rows, err := config.DB.Query(query, sesiID)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil data kehadiran: `+err.Error()+`"}`, 500)
		return
	}
	defer rows.Close()

	type StudentStatus struct {
		NIS       string `json:"nis"`
		Nama      string `json:"nama"`
		Status    string `json:"status"`
		Waktu     string `json:"waktu_absen"`
	}

	var list []StudentStatus
	for rows.Next() {
		var s StudentStatus
		if err := rows.Scan(&s.NIS, &s.Nama, &s.Status, &s.Waktu); err != nil {
			continue
		}
		list = append(list, s)
	}

	// Jika list kosong (tidak ada siswa di kelas tersebut)
	if list == nil {
		list = []StudentStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GetJadwalGuru godoc
// @Summary Jadwal Mengajar Guru
// @Description Menampilkan sesi pembelajaran berdasarkan guru yang sedang login
// @Tags 3. Operasional Guru
// @Produce json
// @Security BearerAuth
// @Success 200 {array} controllers.JadwalGuru "Daftar Jadwal"
// @Router /guru/jadwal [get]
func GetJadwalGuru(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Hanya menerima method GET", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ambil ID dari JWT
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userID := int(userInfo["user_id"].(float64))

	// 2. Cari ID Guru di tabel guru
	var guruID int
	err := config.DB.QueryRow("SELECT id FROM guru WHERE user_id = ?", userID).Scan(&guruID)
	if err != nil {
		http.Error(w, `{"error": "Profil guru tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	// 3. Ambil jadwal Sesi Pembelajaran milik Guru ini saja
	query := `
		SELECT 
			sp.id, k.nama_kelas, mp.nama_mapel, 
			sp.waktu_mulai, sp.waktu_selesai
		FROM sesi_pembelajaran sp
		JOIN kelas k ON sp.kelas_id = k.id
		JOIN mata_pelajaran mp ON sp.mapel_id = mp.id
		WHERE sp.guru_id = ?
		ORDER BY sp.waktu_mulai ASC
	`

	rows, err := config.DB.Query(query, guruID)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil jadwal mengajar"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listSesi []models.SesiResponse

	for rows.Next() {
		var s models.SesiResponse
		if err := rows.Scan(&s.ID, &s.Kelas, &s.MataPelajaran, &s.WaktuMulai, &s.WaktuSelesai); err != nil {
			http.Error(w, `{"error": "Gagal membaca data jadwal"}`, http.StatusInternalServerError)
			return
		}
		listSesi = append(listSesi, s)
	}

	if listSesi == nil {
		listSesi = []models.SesiResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(listSesi)
}

// GenerateQRSesi godoc
// @Summary Tampilkan QR Code (Guru)
// @Description Menghasilkan JWT token beralaku 30 detik untuk diubah jadi gambar QR
// @Tags 3. Operasional Guru
// @Produce json
// @Security BearerAuth
// @Param sesi_id query int true "ID Sesi Pembelajaran"
// @Success 200 {object} controllers.GuruQRResponse "Sukses Membuat QR"
// @Router /guru/generate-qr [get]
// GenerateQRSesi digunakan oleh Guru untuk menampilkan QR di layar proyektor
func GenerateQRSesi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Hanya menerima method GET", http.StatusMethodNotAllowed)
		return
	}

	// Mengambil ID Sesi dari URL, contoh: /guru/generate-qr?sesi_id=1
	sesiID := r.URL.Query().Get("sesi_id")
	if sesiID == "" {
		http.Error(w, `{"error": "ID Sesi tidak ditemukan"}`, http.StatusBadRequest)
		return
	}

	// 1. Validasi Keamanan: Pastikan guru yang login adalah pengajar sesi ini!
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userID := int(userInfo["user_id"].(float64))
	
	var guruID int
	config.DB.QueryRow("SELECT id FROM guru WHERE user_id = ?", userID).Scan(&guruID)

	var validasiSesi int
	err := config.DB.QueryRow("SELECT id FROM sesi_pembelajaran WHERE id = ? AND guru_id = ?", sesiID, guruID).Scan(&validasiSesi)
	if err != nil {
		http.Error(w, `{"error": "Anda tidak berhak membuat QR untuk kelas ini!"}`, http.StatusForbidden)
		return
	}

	// 2. PEMBUATAN TOKEN QR DINAMIS
	claims := jwt.MapClaims{
		"sesi_id": sesiID,
		"tipe":    "qr_absen", // Tanda pengenal bahwa ini bukan token login
		// QR CODE INI HANYA BERLAKU 30 DETIK SEJAK DIBUAT!
		"exp":     time.Now().Add(time.Second * 30).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWT_KEY)
	if err != nil {
		http.Error(w, `{"error": "Gagal membuat QR Code"}`, http.StatusInternalServerError)
		return
	}

	// Di Frontend nanti, string token inilah yang diubah menjadi gambar kotak-kotak QR Code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"message": "QR Code berhasil dibuat, berlaku 30 detik!", "qr_token": "%s"}`, tokenString)
	w.Write([]byte(response))
}


// ExportDataAbsensi godoc
// @Summary Download Laporan Excel
// @Description Otomatis mengunduh file Excel laporan absensi dan mem-backup ke G-Drive
// @Tags 3. Operasional Guru
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} file "File Downloaded"
// @Router /guru/export [get]
// Ekspor Data ke Excel & Google Drive
func ExportDataAbsensi(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil data menggunakan JOIN antara log_kehadiran dan siswa sesuai ERD
	query := `
		SELECT l.id, s.nama_lengkap, l.status_kehadiran, l.waktu_absen 
		FROM log_kehadiran l
		JOIN siswa s ON l.siswa_id = s.id
	`
	rows, err := config.DB.Query(query)
	
	// Penanganan error agar server tidak crash jika query gagal
	if err != nil {
		http.Error(w, "Gagal mengambil data dari DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 2. Buat file Excel
	f := excelize.NewFile()
	sheet := "Sheet1"
	
	// Header kolom disesuaikan
	f.SetCellValue(sheet, "A1", "ID Log")
	f.SetCellValue(sheet, "B1", "Nama Siswa") // Menggunakan nama, bukan ID
	f.SetCellValue(sheet, "C1", "Status Kehadiran")
	f.SetCellValue(sheet, "D1", "Waktu Absen")

	i := 2
	for rows.Next() {
		var id int
		var namaLengkap, status, waktu string
		
		// Proses pemetaan data dari DB ke variabel
		err := rows.Scan(&id, &namaLengkap, &status, &waktu)
		if err != nil {
			fmt.Println("Gagal membaca baris data:", err)
			continue
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", i), id)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i), namaLengkap)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i), status)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", i), waktu)
		i++
	}

	// 3. Simpan sementara di server local
	fileName := "Laporan_Absen_" + time.Now().Format("2006-01-02") + ".xlsx"
	err = f.SaveAs(fileName)
	if err != nil {
		http.Error(w, "Gagal membuat file Excel lokal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Upload ke Google Drive (Background Backup)
	srv, errDrive := helpers.InitDriveService()
	if errDrive == nil {
		folderID := "1akOpzgXyB7r8oO7JSolBnfIGwL46xV__" 
		// Kita abaikan error dari sini agar jika Drive penuh/bermasalah, 
		// Guru tetap bisa mendownload file aslinya
		_ = helpers.UploadToDrive(srv, fileName, folderID)
	}

	// 5. Set HTTP Headers agar Browser Otomatis Mendownload
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")

	// 6. Semburkan file dari memori Go ke Response (Browser Guru)
	err = f.Write(w)
	if err != nil {
		http.Error(w, "Gagal mengirim file ke browser: "+err.Error(), http.StatusInternalServerError)
	}

	// 7. Hapus file sementara di lokal agar hemat storage server
	// Beri jeda kecil agar proses Write(w) selesai sempurna sebelum file dihapus
	time.Sleep(500 * time.Millisecond)
	os.Remove(fileName)
}

// BackupDatabase godoc
// @Summary Backup Database ke Google Drive
// @Description Melakukan dump SQL dan mengunggahnya otomatis ke folder Drive
// @Tags 3. Operasional Guru
// @Produce json
// @Security BearerAuth
// @Success 200 {object} controllers.SuccessMessage "Backup Berhasil Diupload"
// @Router /guru/backup [get]
// BackupDatabase: Dump DB ke SQL lalu upload ke Drive
func BackupDatabase(w http.ResponseWriter, r *http.Request) {
	// 1. Konfigurasi Nama File
	fileName := fmt.Sprintf("Backup_Full_%s.sql", time.Now().Format("2006-01-02_15-04-05"))
	
	// 2. Jalankan perintah mysqldump (Sesuaikan user & nama DB kamu)
	// Format: mysqldump -u [user] -p[pass] [nama_db]
	passDb := fmt.Sprintf("-p%s", os.Getenv("DB_PASSWORD"))
	cmd := exec.Command("mysqldump", "-u", "root", passDb, os.Getenv("DB_NAME"), "--result-file="+fileName)
	err := cmd.Run()
	if err != nil {
		http.Error(w, "Gagal dump database: "+err.Error(), 500)
		return
	}

	// 3. Upload ke Drive
	srv, _ := helpers.InitDriveService()
	folderID := "1OpRprWCk2MgurUclTWdTR43EaoKuOIGH"
	err = helpers.UploadToDrive(srv, fileName, folderID)
	
	// 4. Hapus file lokal setelah upload agar hemat space
	  // Beri jeda 1 detik agar OS/Antivirus melepaskan file ini dari memori
	time.Sleep(1 * time.Second) 
	os.Remove(fileName)
	
	if err != nil {
		http.Error(w, "Gagal upload backup: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Backup Database Berhasil Diupload!"))
}

// RestoreDatabase godoc
// @Summary Restore Database dari Google Drive
// @Description Mengunduh file SQL dari Drive menggunakan File ID dan menimpanya ke database lokal
// @Tags 3. Operasional Guru
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param file_id query string true "Google Drive File ID"
// @Success 200 {object} controllers.SuccessMessage "Database Berhasil Direstore"
// @Router /guru/restore [post]
// RestoreDatabase: Ambil SQL dari Drive lalu timpa ke DB lokal
func RestoreDatabase(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("file_id") // Ambil ID file dari parameter URL
	if fileID == "" {
		http.Error(w, "file_id wajib diisi", 400)
		return
	}

	localFile := "restore_temp.sql"
	srv, _ := helpers.InitDriveService()

	// 1. Download dari Drive
	err := helpers.DownloadFromDrive(srv, fileID, localFile)
	if err != nil {
		http.Error(w, "Gagal download file: "+err.Error(), 500)
		return
	}

	// 2. Eksekusi Restore (Menggunakan mysql command)
	// Format: mysql -u [user] -p[pass] [nama_db] < file.sql
	// Di Go kita pakai pipe untuk input file
	passDb := fmt.Sprintf("-p%s", os.Getenv("DB_PASSWORD"))
	cmd := exec.Command("mysql", "-u", "root", passDb, os.Getenv("DB_NAME"))
	sqlFile, _ := os.Open(localFile)
	cmd.Stdin = sqlFile
	err = cmd.Run()
	
	os.Remove(localFile) // Bersihkan file temp

	if err != nil {
		http.Error(w, "Gagal restore database: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Database Berhasil Direstore ke Kondisi Backup!"))
}