package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"backend-absensi/config"
	"backend-absensi/controllers" // Import folder controllers
	"backend-absensi/middlewares"
	"backend-absensi/helpers"

	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "backend-absensi/docs"
)

// @title API Sistem Absensi Tempat Les
// @version 1.0
// @description Ini adalah dokumentasi REST API untuk aplikasi absensi dengan fitur QR & Geofencing.
// @termsOfService http://swagger.io/terms/

// @contact.name Hafizh
// @contact.email hafizh@example.com

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Masukkan token JWT dengan format: Bearer {token}

func main() {
	// 1. Muat konfigurasi .env paling pertama!
	config.LoadConfig()

	// 2. Menyalakan Database
	config.ConnectDB()

	// 3. Menyiapkan akun awal jika belum ada dan nyalakan scheduler otomatisasi backup
	buatAdminPertama()
	StartBackupScheduler()

	// Rute Publik (Tanpa Satpam)
	http.HandleFunc("/login", controllers.Login)
    http.HandleFunc("/login/siswa", controllers.LoginSiswa)

	// Rute VIP (Dijaga Satpam Middleware)
	http.HandleFunc("/dashboard", middlewares.JWTMiddleware(controllers.DashboardAdmin))
    
	// Rute Khusus Admin
	http.HandleFunc("/admin/dashboard/stats", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetDashboardStats)))
	http.HandleFunc("/admin/backup", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.BackupDatabase)))
    http.HandleFunc("/admin/restore", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.RestoreDatabase)))
    http.HandleFunc("/admin/users/inactive", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetInactiveUsers)))
    http.HandleFunc("/admin/users/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllUsers)))
    http.HandleFunc("/admin/guru/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateGuru)))
    http.HandleFunc("/admin/guru/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteGuru)))
    http.HandleFunc("/admin/guru/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateGuru)))
	http.HandleFunc("/admin/siswa/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateSiswa)))
	http.HandleFunc("/admin/siswa/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllSiswa))) 
	http.HandleFunc("/admin/siswa/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateSiswa)))
	http.HandleFunc("/admin/siswa/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteSiswa)))
    http.HandleFunc("/admin/mapel/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateMapel)))
    http.HandleFunc("/admin/mapel/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllMapel)))
    http.HandleFunc("/admin/mapel/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateMapel)))
    http.HandleFunc("/admin/mapel/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteMapel)))
    http.HandleFunc("/admin/mapel/status", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.ToggleStatusMapel)))
    http.HandleFunc("/admin/kelas/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateKelas)))
    http.HandleFunc("/admin/kelas/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllKelas)))
    http.HandleFunc("/admin/kelas/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateKelas)))
    http.HandleFunc("/admin/kelas/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteKelas)))
    http.HandleFunc("/admin/kelas/siswa", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetSiswaByKelas)))
    http.HandleFunc("/admin/periode/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreatePeriode)))
    http.HandleFunc("/admin/periode/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllPeriode)))
    http.HandleFunc("/admin/periode/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdatePeriode)))
	http.HandleFunc("/admin/periode/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeletePeriode)))
	http.HandleFunc("/admin/siswa-kelas/list", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetKelasBySiswa)))
	http.HandleFunc("/admin/siswa-kelas/assign", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.AssignSiswaToKelas)))
	http.HandleFunc("/admin/siswa-kelas/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateSiswaKelas)))
    http.HandleFunc("/admin/siswa-kelas/remove", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.RemoveSiswaFromKelas)))
	http.HandleFunc("/admin/device/pending", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetPendingDevices)))
    http.HandleFunc("/admin/device/approve", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.ApproveDevice)))
    http.HandleFunc("/admin/device/reject", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.RejectDevice)))
    http.HandleFunc("/admin/siswa/clue", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetSiswaClue)))
	http.HandleFunc("/admin/siswa/reset-password", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.ResetPasswordSiswa)))

	// Rute KHUSUS Super Admin (Manajemen Admin)
    http.HandleFunc("/superadmin/admin/create", middlewares.JWTMiddleware(middlewares.SuperAdminOnly(controllers.CreateAdmin)))
    http.HandleFunc("/superadmin/admin/all", middlewares.JWTMiddleware(middlewares.SuperAdminOnly(controllers.GetAllAdmins)))
    http.HandleFunc("/superadmin/admin/update", middlewares.JWTMiddleware(middlewares.SuperAdminOnly(controllers.UpdateAdmin)))
    http.HandleFunc("/superadmin/admin/toggle", middlewares.JWTMiddleware(middlewares.SuperAdminOnly(controllers.ToggleAdminStatus)))
	http.HandleFunc("/superadmin/users/hard-delete", middlewares.JWTMiddleware(middlewares.SuperAdminOnly(controllers.HardDeleteUser)))
    http.HandleFunc("/superadmin/users/reactivate", middlewares.JWTMiddleware(middlewares.SuperAdminOnly(controllers.ReactivateUser)))


	// Rute Operasional Siswa
	// Perhatikan: Kita pakai satpam SiswaOnly di sini
	http.HandleFunc("/siswa/absen", middlewares.JWTMiddleware(middlewares.SiswaOnly(controllers.CatatAbsen)))
	http.HandleFunc("/siswa/riwayat", middlewares.JWTMiddleware(middlewares.SiswaOnly(controllers.GetRiwayatAbsenSiswa)))
    http.HandleFunc("/siswa/absen/submit", middlewares.JWTMiddleware(controllers.SubmitAbsen))

	// Rute Operasional Guru
	http.HandleFunc("/guru/dashboard/stats", middlewares.JWTMiddleware(controllers.GetGuruStats))
	http.HandleFunc("/guru/jadwal", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.GetJadwalGuru)))
	http.HandleFunc("/guru/absen", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.CatatAbsenManual)))
	http.HandleFunc("/guru/generate-qr", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.GenerateQRSesi)))
    http.HandleFunc("/guru/dashboard/attendance-list", middlewares.JWTMiddleware(controllers.GetSesiSiswaStatus))
    http.HandleFunc("/guru/sesi/init", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.InitOrGetSesi)))
    http.HandleFunc("/guru/mapel", middlewares.JWTMiddleware(controllers.GetMapelForGuru))
    http.HandleFunc("/guru/kelas", middlewares.JWTMiddleware(controllers.GetKelasForGuru))
	http.HandleFunc("/guru/export", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.ExportDataAbsensi)))
    // http.HandleFunc("/guru/backup", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.BackupDatabase)))
	// http.HandleFunc("/guru/restore", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.RestoreDatabase)))
    http.HandleFunc("/guru/forgot-password", controllers.RequestResetPassword)
    http.HandleFunc("/guru/reset-password", controllers.ExecuteResetPassword)

	// Rute Dokumentasi Swagger
    http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Ambil variabel dari env
    frontendURL := os.Getenv("FRONTEND_URL")
    // Kita pecah string dari env menjadi slice string untuk CORS
    allowedOrigins := helpers.ParseCommaSeparated(frontendURL)

	// Jika frontendURL kosong, berikan default localhost agar tidak error saat dev
    if len(allowedOrigins) == 0 {
        allowedOrigins = []string{"http://localhost:5173"}
    }
    // Konfigurasi CORS yang dinamis
    c := cors.New(cors.Options{
        AllowedOrigins:   allowedOrigins, // Mengambil dari env
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        AllowCredentials: true,
    })

    handler := c.Handler(http.DefaultServeMux)

	// Ambil Port dari env (Fly.io biasanya otomatis kasih port lewat env PORT)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // default jika env tidak ada
    }

    log.Println("Server backend berjalan di port 8080...", port)
    // Jalankan server pakai handler yang sudah dibungkus CORS
    if err := http.ListenAndServe(":"+port, handler); err != nil {
	    log.Fatal("Server gagal jalan:", err)
    }
}

// Biarkan fungsi ini di main.go karena ini hanya berjalan sekali saat server menyala
func buatAdminPertama() {
    var role string
    // Cek apakah akun admin_utama sudah ada
    err := config.DB.QueryRow("SELECT role FROM users WHERE username = 'admin_utama'").Scan(&role)
    
    if err != nil {
        // Jika belum ada sama sekali, buat baru sebagai super_admin
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
        _, errInsert := config.DB.Exec("INSERT INTO users (username, password, role, is_active) VALUES (?, ?, ?, 1)", "admin_utama", hashedPassword, "super_admin")
        if errInsert != nil {
            log.Println("Gagal membuat super_admin otomatis:", errInsert)
        } else {
            fmt.Println("-> Akun Super Admin default terbuat! Username: 'admin_utama', Password: 'admin123'")
        }
    } else if role == "admin" {
        // Jika akun admin_utama sudah ada tapi masih admin biasa, UPGRADE otomatis!
        _, errUpdate := config.DB.Exec("UPDATE users SET role = 'super_admin' WHERE username = 'admin_utama'")
        if errUpdate == nil {
            fmt.Println("-> Akun 'admin_utama' berhasil di-upgrade otomatis menjadi Super Admin!")
        }
    }
}

func StartBackupScheduler() {
	// Berjalan setiap 24 jam untuk mengecek apakah ini tanggal 1
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			now := time.Now()
			// Cek jika hari ini tanggal 1
			if now.Day() == 1 {
				fmt.Println("Memulai Backup Otomatis Bulanan...")
				// Kita panggil fungsi backup secara internal
				// Karena BackupDatabase butuh http.ResponseWriter, kita buat fungsi helper
				// Atau simplenya, copy logika backup ke sini tanpa ResponseWriter
				jalankanBackupTanpaHttp()
			}
		}
	}()
}

func jalankanBackupTanpaHttp() {
    // Logika yang sama dengan BackupDatabase tapi tanpa http.Error
    fileName := fmt.Sprintf("Auto_Monthly_Backup_%s.sql", time.Now().Format("2006-01"))
	passDb := fmt.Sprintf("-p%s", os.Getenv("DB_PASSWORD"))
    exec.Command("mysqldump", "-u", "root", passDb, os.Getenv("DB_NAME"), "--result-file="+fileName).Run()
    
    srv, err := helpers.InitDriveService()
    if err == nil {
        helpers.UploadToDrive(srv, fileName, "1OpRprWCk2MgurUclTWdTR43EaoKuOIGH")
    }
    os.Remove(fileName)
    fmt.Println("Backup Otomatis Selesai!")
}
