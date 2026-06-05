package controllers

import (
	"encoding/json"
	"net/http"

	"backend-absensi/config"
	"backend-absensi/models"

	"github.com/golang-jwt/jwt/v5"
)

func CatatAbsen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Hanya menerima method POST", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ambil data pengirim (Siswa) dari Token Login-nya
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userID := int(userInfo["user_id"].(float64))

	var siswaID int
	config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", userID).Scan(&siswaID)

	// 2. Baca data dari HP Siswa
	var req models.AbsenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
		return
	}

	// 3. PEMBONGKARAN & VALIDASI QR CODE
	if req.QRToken == "" {
		http.Error(w, `{"error": "QR Token tidak boleh kosong!"}`, http.StatusBadRequest)
		return
	}

	qrToken, err := jwt.Parse(req.QRToken, func(token *jwt.Token) (interface{}, error) {
		return config.JWT_KEY, nil
	})

	// Jika QR palsu atau sudah lewat dari 30 detik, err akan terisi!
	if err != nil || !qrToken.Valid {
		http.Error(w, `{"error": "QR Code tidak valid atau sudah kedaluwarsa! Silakan minta guru menampilkan QR baru."}`, http.StatusUnauthorized)
		return
	}

	// 4. Ekstrak sesi_id dari dalam QR
    qrClaims, ok := qrToken.Claims.(jwt.MapClaims)
    if !ok || qrClaims["tipe"] != "qr_absen" {
        http.Error(w, `{"error": "Jenis QR Code salah!"}`, http.StatusBadRequest)
        return
    }
    
    sesiID := qrClaims["sesi_id"].(string)

    // ==========================================
    // [TAMBAHAN BARU DI SINI]: Validasi Lintas Kelas
    // ==========================================
    var kelasID int
    errSesi := config.DB.QueryRow("SELECT kelas_id FROM sesi_pembelajaran WHERE id = ?", sesiID).Scan(&kelasID)
    if errSesi != nil {
        http.Error(w, `{"error": "Sesi pembelajaran tidak ditemukan atau sudah tidak valid."}`, http.StatusNotFound)
        return
    }

    var isTerdaftar int
    errCek := config.DB.QueryRow("SELECT COUNT(*) FROM siswa_kelas WHERE siswa_id = ? AND kelas_id = ?", siswaID, kelasID).Scan(&isTerdaftar)

    if errCek != nil || isTerdaftar == 0 {
        http.Error(w, `{"error": "Akses ditolak: Anda terdeteksi tidak terdaftar di kelas ini!"}`, http.StatusForbidden)
        return
    }
    // ==========================================

	// 5. Validasi Status Hadir
	if req.StatusKehadiran != "Hadir" && req.StatusKehadiran != "Izin" && req.StatusKehadiran != "Sakit" {
		http.Error(w, `{"error": "Status kehadiran tidak valid!"}`, http.StatusBadRequest)
		return
	}

	// 6. Simpan secara Paksa sebagai "QR Code"
	query := "INSERT INTO log_kehadiran (sesi_id, siswa_id, status_kehadiran, metode_absen) VALUES (?, ?, ?, 'QR Code')"
	_, err = config.DB.Exec(query, sesiID, siswaID, req.StatusKehadiran)
	
	if err != nil {
		http.Error(w, `{"error": "Gagal mencatat kehadiran. Anda mungkin sudah absen untuk sesi ini."}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Absen berhasil dicatat via QR Code Dinamis!"}`))
}

func CatatAbsenManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Hanya menerima method POST", http.StatusMethodNotAllowed)
		return
	}

	var req models.GuruAbsenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
		return
	}

	// 1. Validasi Status Kehadiran secara ketat
	if req.StatusKehadiran != "Hadir" && req.StatusKehadiran != "Izin" && req.StatusKehadiran != "Sakit" && req.StatusKehadiran != "Alpa" {
		http.Error(w, `{"error": "Status kehadiran tidak valid!"}`, http.StatusBadRequest)
		return
	}

	// 2. Simpan ke Database
	// PERHATIKAN: Kita melakukan hardcode tulisan 'Manual by Guru' langsung di query SQL-nya
	// Ini adalah bentuk perlindungan mutlak dari Backend.
	// Di dalam fungsi CatatAbsenManual
	// ... di dalam fungsi CatatAbsenManual ...

	query := `
	INSERT INTO log_kehadiran (sesi_id, siswa_id, tanggal, status_kehadiran, metode_absen, waktu_absen)
	VALUES (?, ?, ?, ?, 'Manual by Guru', NOW())
	ON DUPLICATE KEY UPDATE
	status_kehadiran = VALUES(status_kehadiran),
	metode_absen = VALUES(metode_absen),
	waktu_absen = NOW()
	`

	_, err := config.DB.Exec(query, req.SesiID, req.SiswaID, req.Tanggal, req.StatusKehadiran)

	// ... lanjut ke pengecekan error ...
	if err != nil {
		// Jika error, biasanya karena ID Sesi atau ID Siswa tidak ada di database (ditolak oleh Foreign Key)
		http.Error(w, `{"error": "Gagal mencatat kehadiran manual. Cek kembali ID Sesi dan ID Siswa."}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Absen manual berhasil dicatat oleh Guru!"}`))
}

// GetRiwayatAbsenSiswa godoc
// @Summary Riwayat Absensi Siswa
// @Description Menampilkan data historis absensi milik siswa yang sedang login
// @Tags 2. Operasional Siswa
// @Produce json
// @Security BearerAuth
// @Success 200 {array} controllers.RiwayatAbsenSiswa "Daftar Riwayat"
// @Router /siswa/riwayat [get]
func GetRiwayatAbsenSiswa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Hanya menerima method GET", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ambil ID Akun dari JWT (Gelang Tiket)
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userID := int(userInfo["user_id"].(float64))

	// 2. Cari ID Siswa dari tabel siswa
	var siswaID int
	err := config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", userID).Scan(&siswaID)
	if err != nil {
		http.Error(w, `{"error": "Profil siswa tidak ditemukan"}`, http.StatusNotFound)
		return
	}

	// 3. Query Super JOIN untuk mengambil riwayat yang lengkap dan mudah dibaca
	query := `
		SELECT 
			DATE_FORMAT(sp.waktu_mulai, '%Y-%m-%d') as tanggal, 
			mp.nama_mapel, 
			g.nama_lengkap as guru_pengajar, 
			lk.status_kehadiran
		FROM log_kehadiran lk
		JOIN sesi_pembelajaran sp ON lk.sesi_id = sp.id
		JOIN mata_pelajaran mp ON sp.mapel_id = mp.id
		JOIN guru g ON sp.guru_id = g.id
		WHERE lk.siswa_id = ?
		ORDER BY sp.waktu_mulai DESC
	`

	rows, err := config.DB.Query(query, siswaID)
	if err != nil {
		http.Error(w, `{"error": "Gagal mengambil riwayat absensi"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var riwayat []models.RiwayatAbsenResponse

	for rows.Next() {
		var r models.RiwayatAbsenResponse
		if err := rows.Scan(&r.Tanggal, &r.MataPelajaran, &r.GuruPengajar, &r.StatusKehadiran); err != nil {
			http.Error(w, `{"error": "Gagal membaca data riwayat"}`, http.StatusInternalServerError)
			return
		}
		riwayat = append(riwayat, r)
	}

	if riwayat == nil {
		riwayat = []models.RiwayatAbsenResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(riwayat)
}
