package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"backend-absensi/config"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

// ImportExcelSiswa menerima file Excel template dan melakukan registrasi masal
func ImportExcelSiswa(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
		return
	}

	// 1. Terima File dari Request Form-Data
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, `{"error": "Gagal membaca form atau file terlalu besar"}`, http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file_excel")
	if err != nil {
		http.Error(w, `{"error": "File Excel tidak ditemukan di request"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		http.Error(w, `{"error": "Format file tidak valid. Pastikan file berformat .xlsx"}`, http.StatusBadRequest)
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, `{"error": "Gagal membaca baris data di Sheet pertama"}`, http.StatusInternalServerError)
		return
	}

	// --- PERSIAPAN BIKIN FILE EXCEL HASIL GENERATE ---
	resultFile := excelize.NewFile()
	resultSheet := "Akun_Siswa"
	resultFile.SetSheetName("Sheet1", resultSheet)
	// Set Header Kolom
	resultFile.SetCellValue(resultSheet, "A1", "Nama Lengkap")
	resultFile.SetCellValue(resultSheet, "B1", "Nama Sekolah")
	resultFile.SetCellValue(resultSheet, "C1", "Username")
	resultFile.SetCellValue(resultSheet, "D1", "Password")
	resultRow := 2 // Data dimulai dari baris ke-2

	sukses := 0
	gagal := 0
	duplikat := 0

	tx, err := config.DB.Begin()
	if err != nil {
		http.Error(w, `{"error": "Gagal memulai database transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// 2. Looping Baris per Baris
	for i, row := range rows {
		// Skip baris 1 (Header)
		if i == 0 { continue }
		if len(row) < 1 { continue }

		namaLengkap := strings.TrimSpace(row[0])
		if namaLengkap == "" { continue }

		namaSekolah := "-"
		if len(row) > 1 && strings.TrimSpace(row[1]) != "" {
			namaSekolah = strings.TrimSpace(row[1])
		}

		clueKataKunci := "Default"
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			clueKataKunci = strings.TrimSpace(row[2])
		}

		kataKunci := "-"
		if len(row) > 3 && strings.TrimSpace(row[3]) != "" {
			kataKunci = strings.TrimSpace(row[3])
		}

		// --- PENGECEKAN DUPLIKASI DATA (Tanpa LOWER) ---
		var count int
		checkQuery := `SELECT COUNT(*) FROM siswa WHERE nama_lengkap = ? AND nama_sekolah = ?`
		errCheck := tx.QueryRow(checkQuery, namaLengkap, namaSekolah).Scan(&count)

		if errCheck != nil {
			gagal++
			continue
		}

		if count > 0 {
			duplikat++
			continue
		}

		// --- A. LOGIKA GENERATE USERNAME ---
		baseUsername := strings.ToLower(strings.ReplaceAll(namaLengkap, " ", ""))
		if len(baseUsername) > 15 {
			baseUsername = baseUsername[:15]
		}
		username := fmt.Sprintf("%s%03d", baseUsername, rand.Intn(999))

		// --- B. LOGIKA PASSWORD LOGIN AWAL ---
		cleanName := strings.ToLower(strings.ReplaceAll(namaLengkap, " ", ""))
		rawPassword := ""

		if len(cleanName) >= 6 {
			rawPassword = cleanName[:3] + cleanName[len(cleanName)-3:]
		} else {
			rawPassword = cleanName + cleanName
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)

		// --- C. EKSEKUSI INSERT KE DB ---
		queryUser := `INSERT INTO users (username, password, role, is_active, created_at) VALUES (?, ?, 'siswa', 1, NOW())`
		res, errInsertUser := tx.Exec(queryUser, username, hashedPassword)

		if errInsertUser != nil {
			gagal++
			continue
		}

		userID, _ := res.LastInsertId()

		querySiswa := `INSERT INTO siswa (user_id, nama_sekolah, nama_lengkap, label_kata_kunci, kata_kunci) VALUES (?, ?, ?, ?, ?)`
		_, errInsertSiswa := tx.Exec(querySiswa, userID, namaSekolah, namaLengkap, clueKataKunci, kataKunci)

		if errInsertSiswa != nil {
			gagal++
			continue
		}

		// --- D. CATAT KE FILE EXCEL HASIL (Hanya yang sukses) ---
		resultFile.SetCellValue(resultSheet, fmt.Sprintf("A%d", resultRow), namaLengkap)
		resultFile.SetCellValue(resultSheet, fmt.Sprintf("B%d", resultRow), namaSekolah)
		resultFile.SetCellValue(resultSheet, fmt.Sprintf("C%d", resultRow), username)
		resultFile.SetCellValue(resultSheet, fmt.Sprintf("D%d", resultRow), rawPassword) // Password ASLI, belum di-hash
		resultRow++
		
		sukses++
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"error": "Gagal menyimpan perubahan ke database"}`, http.StatusInternalServerError)
		return
	}

	// --- 3. KEMBALIKAN RESPONSE SESUAI KONDISI ---

	// Kondisi 1: Jika tidak ada data yang masuk (semua duplikat/gagal) -> Kembalikan JSON
	if sukses == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"message": "Selesai. Tidak ada siswa baru yang ditambahkan.", "sukses": %d, "gagal": %d, "duplikat": %d}`, sukses, gagal, duplikat)))
		return
	}

	// Kondisi 2: Jika ada yang sukses -> Kembalikan FILE EXCEL
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="Hasil_Akun_Siswa.xlsx"`)
	
	// Sisipkan rekapitulasi data di dalam Custom HTTP Headers agar bisa dibaca Svelte
	w.Header().Set("X-Import-Success", fmt.Sprintf("%d", sukses))
	w.Header().Set("X-Import-Duplicate", fmt.Sprintf("%d", duplikat))
	w.Header().Set("X-Import-Failed", fmt.Sprintf("%d", gagal))
	// Expose header ini wajib agar Browser mengizinkan Svelte membacanya (CORS)
	w.Header().Set("Access-Control-Expose-Headers", "X-Import-Success, X-Import-Duplicate, X-Import-Failed")

	if err := resultFile.Write(w); err != nil {
		http.Error(w, `{"error": "Gagal men-generate file hasil"}`, http.StatusInternalServerError)
		return
	}
}