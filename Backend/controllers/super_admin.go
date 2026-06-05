package controllers

import (
	"fmt"
    "encoding/json"
    "net/http"

    "backend-absensi/config"
    "golang.org/x/crypto/bcrypt"
)

type AdminRequest struct {
    Username string `json:"username"`
    Password string `json:"password"` // Kosongkan saat update jika tidak ingin ganti password
}

func CreateAdmin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, `{"error": "Method tidak diizinkan"}`, http.StatusMethodNotAllowed)
        return
    }

    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "Format data tidak valid"}`, http.StatusBadRequest)
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengenkripsi password"}`, http.StatusInternalServerError)
        return
    }

    // Role di-hardcode sebagai "admin" biasa
    query := `INSERT INTO users (username, password, role, is_active) VALUES (?, ?, 'admin', 1)`
    _, err = config.DB.Exec(query, req.Username, hashedPassword)
    if err != nil {
        http.Error(w, `{"error": "Username mungkin sudah terdaftar."}`, http.StatusConflict)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(`{"message": "Akun Admin baru berhasil dibuat!"}`))
}

func GetAllAdmins(w http.ResponseWriter, r *http.Request) {
    // Ambil semua user dengan role 'admin' (tidak termasuk super_admin)
    rows, err := config.DB.Query("SELECT id, username, is_active FROM users WHERE role = 'admin'")
    if err != nil {
        http.Error(w, `{"error": "Gagal mengambil data admin"}`, http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type AdminResponse struct {
        ID       int    `json:"id"`
        Username string `json:"username"`
        IsActive int    `json:"is_active"`
    }

    var list []AdminResponse
    for rows.Next() {
        var a AdminResponse
        if err := rows.Scan(&a.ID, &a.Username, &a.IsActive); err == nil {
            list = append(list, a)
        }
    }
    if list == nil {
        list = []AdminResponse{}
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(list)
}

func UpdateAdmin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, `{"error": "Hanya menerima method PUT"}`, http.StatusMethodNotAllowed)
        return
    }

    adminID := r.URL.Query().Get("id")
    if adminID == "" {
        http.Error(w, `{"error": "ID Admin tidak ditemukan di URL"}`, http.StatusBadRequest)
        return
    }

    var req AdminRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, `{"error": "Format data salah"}`, http.StatusBadRequest)
        return
    }

    if req.Password != "" {
        // Update username & password
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        _, err := config.DB.Exec("UPDATE users SET username = ?, password = ? WHERE id = ? AND role = 'admin'", req.Username, hashedPassword, adminID)
        if err != nil {
            http.Error(w, `{"error": "Gagal update atau username sudah dipakai"}`, http.StatusInternalServerError)
            return
        }
    } else {
        // Update username saja
        _, err := config.DB.Exec("UPDATE users SET username = ? WHERE id = ? AND role = 'admin'", req.Username, adminID)
        if err != nil {
            http.Error(w, `{"error": "Gagal update atau username sudah dipakai"}`, http.StatusInternalServerError)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Data Admin berhasil diperbarui!"}`))
}

func ToggleAdminStatus(w http.ResponseWriter, r *http.Request) {
    adminID := r.URL.Query().Get("id")
    if adminID == "" {
        http.Error(w, `{"error": "ID Admin tidak valid"}`, http.StatusBadRequest)
        return
    }

    // Toggle status aktif (1 jadi 0, 0 jadi 1)
    _, err := config.DB.Exec("UPDATE users SET is_active = NOT is_active WHERE id = ? AND role = 'admin'", adminID)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengubah status admin."}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Status Admin berhasil diubah!"}`))
}

func HardDeleteUser(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, `{"error": "Hanya menerima method DELETE"}`, http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, `{"error": "ID User wajib diisi"}`, http.StatusBadRequest)
        return
    }

    // 1. Cari tahu dulu role user-nya
    var role string
    err := config.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
    if err != nil {
        http.Error(w, `{"error": "User tidak ditemukan di database"}`, http.StatusNotFound)
        return
    }

    if role == "super_admin" {
        http.Error(w, `{"error": "Dilarang keras menghapus akun Super Admin!"}`, http.StatusForbidden)
        return
    }

    // 2. Mulai Operasi Sapu Jagat (Transaksi)
    tx, err := config.DB.Begin()
    if err != nil {
        http.Error(w, `{"error": "Gagal memulai pembersihan database"}`, http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    // 3. Hapus data binding perangkat
    tx.Exec("DELETE FROM user_devices WHERE user_id = ?", userID)

    // 4. Hapus data berdasarkan rantai Role-nya
    if role == "siswa" {
        var siswaID int
        errSiswa := tx.QueryRow("SELECT id FROM siswa WHERE user_id = ?", userID).Scan(&siswaID)
        if errSiswa == nil {
            tx.Exec("DELETE FROM log_kehadiran WHERE siswa_id = ?", siswaID)
            tx.Exec("DELETE FROM siswa_kelas WHERE siswa_id = ?", siswaID)
            tx.Exec("DELETE FROM siswa WHERE id = ?", siswaID)
        }
    } else if role == "guru" {
        var guruID int
        errGuru := tx.QueryRow("SELECT id FROM guru WHERE user_id = ?", userID).Scan(&guruID)
        if errGuru == nil {
            tx.Exec("DELETE FROM log_kehadiran WHERE sesi_id IN (SELECT id FROM sesi_pembelajaran WHERE guru_id = ?)", guruID)
            tx.Exec("DELETE FROM sesi_pembelajaran WHERE guru_id = ?", guruID)
            tx.Exec("DELETE FROM guru WHERE id = ?", guruID)
        }
    } else if role == "admin" {
        // Admin tidak punya relasi ke tabel detail, jadi cukup log saja dan biarkan lanjut ke tahap 5
        fmt.Printf("Memproses penghapusan permanen untuk akun Admin ID: %s\n", userID)
    } else {
        // Tolak jika role tidak dikenali untuk keamanan ekstra
        http.Error(w, `{"error": "Role user tidak valid untuk dihapus"}`, http.StatusBadRequest)
        return
    }

    // 5. Terakhir, tebang akar utamanya di tabel users
    _, err = tx.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        http.Error(w, fmt.Sprintf(`{"error": "Gagal mengeksekusi penghapusan permanen: %s"}`, err.Error()), http.StatusInternalServerError)
        return
    }

    // 6. Konfirmasi perubahan permanen ke database
    tx.Commit()

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Beres! User beserta seluruh jejak datanya berhasil dihapus permanen."}`))
}

func ReactivateUser(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, `{"error": "Hanya menerima method PUT"}`, http.StatusMethodNotAllowed)
        return
    }

    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, `{"error": "ID User wajib diisi"}`, http.StatusBadRequest)
        return
    }

    // Pastikan user tersebut memang ada di database sebelum diaktifkan
    var role string
    err := config.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
    if err != nil {
        http.Error(w, `{"error": "User tidak ditemukan."}`, http.StatusNotFound)
        return
    }

    // Eksekusi kebangkitan! (Set is_active menjadi 1)
    _, err = config.DB.Exec("UPDATE users SET is_active = 1 WHERE id = ?", userID)
    if err != nil {
        http.Error(w, `{"error": "Gagal mengaktifkan kembali user ini."}`, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(fmt.Sprintf(`{"message": "Berhasil! Akun %s tersebut telah diaktifkan kembali dan bisa login sekarang."}`, role)))
}

func GetActivityLogs(w http.ResponseWriter, r *http.Request) {
    // Kita Join dengan tabel users untuk mendapatkan username pelakunya
    query := `
    SELECT a.id, u.username, a.role, a.action, a.deskripsi, a.created_at
    FROM activity_logs a
    JOIN users u ON a.user_id = u.id
    ORDER BY a.created_at DESC
    LIMIT 200 -- Dibatasi 200 log terbaru agar query tidak berat
    `
    rows, err := config.DB.Query(query)
    if err != nil {
        http.Error(w, "Gagal menarik data log: "+err.Error(), 500)
        return
    }
    defer rows.Close()

    type LogResponse struct {
        ID        int    `json:"id"`
        Username  string `json:"username"`
        Role      string `json:"role"`
        Action    string `json:"action"`
        Deskripsi string `json:"deskripsi"`
        CreatedAt string `json:"created_at"`
    }

    var list []LogResponse
    for rows.Next() {
        var l LogResponse
        if err := rows.Scan(&l.ID, &l.Username, &l.Role, &l.Action, &l.Deskripsi, &l.CreatedAt); err == nil {
            list = append(list, l)
        }
    }

    // Kembalikan array kosong jika log masih belum ada (mencegah null di Svelte)
    if list == nil {
        list = []LogResponse{}
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(list)
}