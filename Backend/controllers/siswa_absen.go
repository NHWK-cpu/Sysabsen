package controllers

import (
	"encoding/json"
	"net/http"
	"fmt"
	"os"
	"strconv"
	"log"

	"backend-absensi/config"
	"backend-absensi/helpers"
	"github.com/golang-jwt/jwt/v5"
)

// SubmitAbsen godoc
// @Summary Submit Absensi Geofencing (Siswa)
// @Description Mengirim token QR dinamis dan koordinat GPS Siswa untuk divalidasi
// @Tags 2. Operasional Siswa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body controllers.AbsenRequest true "Kirim Koordinat dan Token"
// @Success 200 {object} controllers.SuccessMessage "Posisi tervalidasi, Absen Sukses"
// @Failure 403 {object} controllers.ErrorMessage "Di luar jangkauan radius / QR Kadaluwarsa"
// @Router /siswa/absen/submit [post]
func SubmitAbsen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		SesiID    int     `json:"sesi_id"`
		QRToken   string  `json:"qr_token"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
		return
	}

	// 1. Ambil Data Siswa dari JWT
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userID := int(userInfo["user_id"].(float64))

	var siswaID int
	err := config.DB.QueryRow("SELECT id FROM siswa WHERE user_id = ?", userID).Scan(&siswaID)
	if err != nil {
		http.Error(w, `{"error": "Profil siswa tidak ditemukan"}`, http.StatusForbidden)
		return
	}

	// 2. VALIDASI GEOFENCING (Tarik data dari .env)
	// Gunakan nilai default jika di .env lupa diset agar server tidak crash
	titikLat, errLat := strconv.ParseFloat(os.Getenv("TITIK_LES_LAT"), 64)
	titikLon, errLon := strconv.ParseFloat(os.Getenv("TITIK_LES_LON"), 64)
	radiusToleransi, errRad := strconv.ParseFloat(os.Getenv("RADIUS_TOLERANSI"), 64)

	// Jika parsing gagal, set nilai default darurat (Opsional, tapi aman)
	if errLat != nil || errLon != nil || errRad != nil {
		titikLat = -7.250445
		titikLon = 112.768845
		radiusToleransi = 50.0
	}

	jarak := helpers.HaversineDistance(titikLat, titikLon, input.Latitude, input.Longitude)
	
	if jarak > radiusToleransi {
		http.Error(w, `{"error": "Anda berada di luar jangkauan tempat les. Jarak Anda: `+fmt.Sprintf("%.1f", jarak)+` meter"}`, http.StatusForbidden)
		return
	}

	// 3. VALIDASI QR CODE DINAMIS
	tokenQR, err := jwt.Parse(input.QRToken, func(token *jwt.Token) (interface{}, error) {
		// Pastikan metode enkripsinya sama
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode enkripsi tidak valid")
		}
		// Gunakan Secret Key yang sama dengan yang di guru.go
		return config.JWT_KEY, nil 
	})

	if err != nil {
		// TAMBAHKAN BARIS INI: Biar error aslinya tercetak di journalctl VPS
		log.Printf("GAGAL VERIFIKASI QR DARI USER %d: %v", userID, err)

		http.Error(w, `{"error": "QR Code tidak valid atau sudah kedaluwarsa. Silakan minta Guru menampilkan QR baru."}`, http.StatusForbidden)
		return
	}

	if claims, ok := tokenQR.Claims.(jwt.MapClaims); ok && tokenQR.Valid {
        // A. Pastikan ini benar-benar token QR...
        if claims["tipe"] != "qr_absen" {
            http.Error(w, `{"error": "Tipe QR Code salah!"}`, http.StatusForbidden)
            return
        }

        // B. Pastikan QR yang di-scan sesuai dengan kelas...
        qrSesiID := fmt.Sprintf("%v", claims["sesi_id"]) 
        inputSesiID := fmt.Sprintf("%d", input.SesiID)

        if qrSesiID != inputSesiID {
            http.Error(w, `{"error": "QR Code ini untuk kelas/sesi yang berbeda!"}`, http.StatusForbidden)
            return
        }
    } else {
        http.Error(w, `{"error": "Data QR Code rusak"}`, http.StatusForbidden)
        return
    }

    // ==========================================
    // Validasi Lintas Kelas
    // ==========================================
    var kelasID int
    errSesi := config.DB.QueryRow("SELECT kelas_id FROM sesi_pembelajaran WHERE id = ?", input.SesiID).Scan(&kelasID)
    if errSesi != nil {
        http.Error(w, `{"error": "Sesi pembelajaran tidak ditemukan atau sudah tidak valid."}`, http.StatusNotFound)
        return
    }

    var isTerdaftar int
    // Menggunakan siswaID dan kelasID
    errCek := config.DB.QueryRow("SELECT COUNT(*) FROM siswa_kelas WHERE siswa_id = ? AND kelas_id = ?", siswaID, kelasID).Scan(&isTerdaftar)

    if errCek != nil || isTerdaftar == 0 {
        http.Error(w, `{"error": "Akses ditolak: Anda terdeteksi tidak terdaftar di kelas ini!"}`, http.StatusForbidden)
        return
    }
    // ==========================================

	// 4. CEK APAKAH SUDAH ABSEN SEBELUMNYA
	var sudahAbsen int
	config.DB.QueryRow("SELECT COUNT(*) FROM log_kehadiran WHERE sesi_id = ? AND siswa_id = ?", input.SesiID, siswaID).Scan(&sudahAbsen)
	if sudahAbsen > 0 {
		http.Error(w, `{"error": "Anda sudah melakukan absen untuk sesi ini"}`, http.StatusConflict)
		return
	}

	// 5. SIMPAN KE LOG KEHADIRAN (Sesuai ERD)
	_, err = config.DB.Exec(
		"INSERT INTO log_kehadiran (sesi_id, siswa_id, status_kehadiran, metode_absen, tanggal) VALUES (?, ?, 'Hadir', 'qr_scan', CURDATE())",
		input.SesiID, siswaID,
	)

	if err != nil {
		http.Error(w, `{"error": "Gagal menyimpan absen: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Absen berhasil! Posisi Anda tervalidasi."}`))
}
