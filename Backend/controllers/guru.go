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

	// 3. Simpan sementara di server
	fileName := "Laporan_Absen_" + time.Now().Format("2006-01-02") + ".xlsx"
	err = f.SaveAs(fileName)
	if err != nil {
		http.Error(w, "Gagal membuat file Excel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Upload ke Drive (Pastikan helper Google Drive sudah disetup)
	srv, err := helpers.InitDriveService()
	if err != nil {
		http.Error(w, "Gagal inisialisasi Drive API: "+err.Error(), http.StatusInternalServerError)
		return
	}

	folderID := "1akOpzgXyB7r8oO7JSolBnfIGwL46xV__" 
	err = helpers.UploadToDrive(srv, fileName, folderID)
	if err != nil {
		http.Error(w, "Gagal upload ke Google Drive: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Berhasil! File Laporan Absensi sudah dikirim ke Google Drive."))
}

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