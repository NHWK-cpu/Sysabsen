package controllers

import (
	"encoding/json"
	"net/http"

	"backend-absensi/config"
	"backend-absensi/models" // Import package models yang baru
)

// ==========================================
// FITUR PERIODE BELAJAR
// ==========================================

func CreatePeriode(w http.ResponseWriter, r *http.Request) {
	var p models.PeriodeBelajar
	json.NewDecoder(r.Body).Decode(&p)
	_, err := config.DB.Exec("INSERT INTO periode_belajar (tahun_ajaran, status_aktif) VALUES (?, ?)", p.TahunAjar, p.StatusAktif)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Periode berhasil dibuat"))
}

func GetAllPeriode(w http.ResponseWriter, r *http.Request) {
    // Gunakan err, jangan diabaikan pakai _
    rows, err := config.DB.Query("SELECT id, tahun_ajaran, status_aktif FROM periode_belajar")
    if err != nil {
        http.Error(w, "Gagal query periode: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var list []models.PeriodeBelajar
    for rows.Next() {
        var p models.PeriodeBelajar
        if err := rows.Scan(&p.ID, &p.TahunAjar, &p.StatusAktif); err != nil {
            continue
		}
        list = append(list, p)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(list)
}

func DeletePeriode(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	// Database akan menolak jika periode ini masih dipakai oleh tabel 'kelas' (FK Constraint)
	_, err := config.DB.Exec("DELETE FROM periode_belajar WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Gagal hapus periode (masih ada kelas yang terhubung): "+err.Error(), 500)
		return
	}
	w.Write([]byte("Periode berhasil dihapus"))
}

// ==========================================
// FITUR PIVOT SISWA-KELAS (Assign Siswa ke Kelas)
// ==========================================

func AssignSiswaToKelas(w http.ResponseWriter, r *http.Request) {
	var sk models.SiswaKelas
	json.NewDecoder(r.Body).Decode(&sk)
	_, err := config.DB.Exec("INSERT INTO siswa_kelas (siswa_id, kelas_id) VALUES (?, ?)", sk.SiswaID, sk.KelasID)
	if err != nil {
		http.Error(w, "Gagal assign siswa: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Siswa berhasil dimasukkan ke kelas"))
}
func UpdateSiswaKelas(w http.ResponseWriter, r *http.Request) {
	var sk models.SiswaKelas
	json.NewDecoder(r.Body).Decode(&sk)
	
	// Digunakan untuk pindah kelas (mutasi)
	_, err := config.DB.Exec("UPDATE siswa_kelas SET kelas_id = ? WHERE siswa_id = ?", sk.KelasID, sk.SiswaID)
	if err != nil {
		http.Error(w, "Gagal update kelas siswa: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Data kelas siswa berhasil diperbarui (Mutasi Berhasil)"))
}

func RemoveSiswaFromKelas(w http.ResponseWriter, r *http.Request) {
	siswaID := r.URL.Query().Get("siswa_id")
	kelasID := r.URL.Query().Get("kelas_id")
	
	_, err := config.DB.Exec("DELETE FROM siswa_kelas WHERE siswa_id = ? AND kelas_id = ?", siswaID, kelasID)
	if err != nil {
		http.Error(w, "Gagal mengeluarkan siswa dari kelas: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Siswa berhasil dikeluarkan dari kelas tersebut"))
}

// ==========================================
// FITUR MATA PELAJARAN (MAPEL)
// ==========================================

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
	rows, err := config.DB.Query("SELECT id, nama_mapel FROM mata_pelajaran")
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listMapel []models.Mapel // Slice menggunakan models.Mapel
	for rows.Next() {
		var m models.Mapel
		if err := rows.Scan(&m.ID, &m.NamaMapel); err != nil {
			continue
		}
		listMapel = append(listMapel, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listMapel)
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
	_, err := config.DB.Exec("DELETE FROM mata_pelajaran WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Gagal hapus (mungkin masih digunakan di tabel lain): "+err.Error(), 500)
		return
	}
	w.Write([]byte("Mapel berhasil dihapus"))
}

// ==========================================
// FITUR KELAS
// ==========================================

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
	// Query disesuaikan untuk mengambil periode_id
	rows, err := config.DB.Query("SELECT id, periode_id, nama_kelas FROM kelas")
	if err != nil {
		http.Error(w, "Gagal mengambil data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var listKelas []models.Kelas
	for rows.Next() {
		var k models.Kelas
		// Urutan Scan harus sama persis dengan urutan SELECT
		if err := rows.Scan(&k.ID, &k.PeriodeID, &k.NamaKelas); err != nil {
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
	_, err := config.DB.Exec("DELETE FROM kelas WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Gagal hapus kelas: "+err.Error(), 500)
		return
	}
	w.Write([]byte("Kelas berhasil dihapus"))
}