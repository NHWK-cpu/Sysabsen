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
		TotalUsers       int `json:"total_users"`
		ActiveUsers      int `json:"active_users"`
		InactiveUsers    int `json:"inactive_users"`
		PendingApprovals int `json:"pending_approvals"`
	}

	// 1. Total Users (Semua akun selain admin)
	_ = config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role != 'admin'").Scan(&stats.TotalUsers)

	// 2. Active Users (Login dalam 90 hari terakhir)
	_ = config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE last_login >= DATE_SUB(NOW(), INTERVAL 90 DAY) AND role != 'admin'").Scan(&stats.ActiveUsers)

	// 3. Inactive Users (Tidak login > 90 hari, atau belum pernah login sama sekali)
	_ = config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE (last_login < DATE_SUB(NOW(), INTERVAL 90 DAY) OR last_login IS NULL) AND role != 'admin'").Scan(&stats.InactiveUsers)

	// 4. Pending Approvals (Antrean Perangkat dari Device Binding ERD kita)
	_ = config.DB.QueryRow("SELECT COUNT(*) FROM user_devices WHERE status = 'pending'").Scan(&stats.PendingApprovals)

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
		AND u.role != 'admin'`

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
	query := `
		SELECT u.id, 
		       COALESCE(s.nama_lengkap, g.nama_lengkap, 'User Baru') as nama_lengkap, 
		       u.username, 
		       u.role, 
		       IFNULL(u.last_login, '-') as last_login,
		       IF(u.last_login >= DATE_SUB(NOW(), INTERVAL 90 DAY), 'Active', 'Inactive') as status
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

	var list []models.UserDashboardRow
	for rows.Next() {
		var u models.UserDashboardRow
		err := rows.Scan(&u.ID, &u.NamaLengkap, &u.Username, &u.Role, &u.LastLogin, &u.Status)
		if err != nil {
			continue
		}
		list = append(list, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}