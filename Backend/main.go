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
    http.HandleFunc("/admin/users/inactive", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetInactiveUsers)))
    http.HandleFunc("/admin/users/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllUsers)))
	http.HandleFunc("/admin/siswa", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateSiswa)))
	http.HandleFunc("/admin/siswa/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllSiswa))) 
	http.HandleFunc("/admin/siswa/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateSiswa)))
	http.HandleFunc("/admin/siswa/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteSiswa)))
    http.HandleFunc("/admin/mapel/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateMapel)))
    http.HandleFunc("/admin/mapel/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllMapel)))
    http.HandleFunc("/admin/mapel/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateMapel)))
    http.HandleFunc("/admin/mapel/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteMapel)))
    http.HandleFunc("/admin/kelas/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreateKelas)))
    http.HandleFunc("/admin/kelas/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllKelas)))
    http.HandleFunc("/admin/kelas/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateKelas)))
    http.HandleFunc("/admin/kelas/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeleteKelas)))
    http.HandleFunc("/admin/periode/create", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.CreatePeriode)))
    http.HandleFunc("/admin/periode/all", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetAllPeriode)))
	http.HandleFunc("/admin/periode/delete", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.DeletePeriode)))
	http.HandleFunc("/admin/siswa-kelas/assign", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.AssignSiswaToKelas)))
	http.HandleFunc("/admin/siswa-kelas/update", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.UpdateSiswaKelas)))
    http.HandleFunc("/admin/siswa-kelas/remove", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.RemoveSiswaFromKelas)))
	http.HandleFunc("/admin/device/pending", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.GetPendingDevices)))
    http.HandleFunc("/admin/device/approve", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.ApproveDevice)))
    http.HandleFunc("/admin/device/reject", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.RejectDevice)))
	http.HandleFunc("/admin/siswa/reset-password", middlewares.JWTMiddleware(middlewares.AdminOnly(controllers.ResetPasswordSiswa)))


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
	http.HandleFunc("/guru/export", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.ExportDataAbsensi)))
    http.HandleFunc("/guru/backup", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.BackupDatabase)))
	http.HandleFunc("/guru/restore", middlewares.JWTMiddleware(middlewares.GuruOnly(controllers.RestoreDatabase)))
    http.HandleFunc("/guru/forgot-password", controllers.RequestResetPassword)
    http.HandleFunc("/guru/reset-password", controllers.ExecuteResetPassword)

	// Rute Dokumentasi Swagger
    http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// 4. Menyalakan Server
	fmt.Println("Server berjalan di port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Biarkan fungsi ini di main.go karena ini hanya berjalan sekali saat server menyala
func buatAdminPertama() {
	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		
		_, err := config.DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "admin_utama", hashedPassword, "admin")
		if err != nil {
			log.Println("Gagal membuat admin otomatis:", err)
		} else {
			fmt.Println("-> Akun default terbuat! Username: 'admin_utama', Password: 'admin123'")
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