package controllers

// File ini HANYA digunakan untuk dokumentasi Swagger agar contoh response akurat.
// File ini tidak dieksekusi atau mempengaruhi logika backend sama sekali.

type SuccessMessage struct {
	Message string `json:"message" example:"Aksi berhasil dilakukan!"`
}

type ErrorMessage struct {
	Error string `json:"error" example:"Terjadi kesalahan: data tidak valid"`
}

// -- MOCK LOGIN --
type LoginSiswaRequest struct {
	Username    string `json:"username" example:"siswa01"`
	Password    string `json:"password" example:"password123"`
	DeviceToken string `json:"device_token" example:"hp-hafizh-xyz-123"`
}

type LoginSiswaResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Login berhasil!"`
	Data    struct {
		Role  string `json:"role" example:"siswa"`
		Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6Ikp..."`
	} `json:"data"`
}

// -- MOCK GURU --
type GuruQRResponse struct {
	Message string `json:"message" example:"QR Code berhasil dibuat, berlaku 30 detik!"`
	QRToken string `json:"qr_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6..."`
}

type GuruStatsResponse struct {
	TotalStudents  int     `json:"total_students" example:"30"`
	PresentToday   int     `json:"present_today" example:"28"`
	AbsentToday    int     `json:"absent_today" example:"2"`
	AttendanceRate float64 `json:"attendance_rate" example:"93.33"`
}

type StudentStatusList struct {
	NIS    string `json:"nis" example:"22010101"`
	Nama   string `json:"nama" example:"Hafizh Junior"`
	Status string `json:"status" example:"hadir"`
	Waktu  string `json:"waktu_absen" example:"2026-04-08 09:15:00"`
}

// -- MOCK SISWA --
type AbsenRequest struct {
	SesiID    int     `json:"sesi_id" example:"1"`
	QRToken   string  `json:"qr_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	Latitude  float64 `json:"latitude" example:"-7.250446"`
	Longitude float64 `json:"longitude" example:"112.768846"`
}

// -- MOCK AUTH GURU --
type ForgotPasswordReq struct {
	Email string `json:"email" example:"guru@tempatles.com"`
}

type ResetPasswordReq struct {
	Token       string `json:"token" example:"817cdc18e67140e57f62dd6c80fd52529..."`
	NewPassword string `json:"new_password" example:"PasswordBaru123!"`
}

// -- MOCK ADMIN & DEVICE --
type LoginRequest struct {
	Username string `json:"username" example:"admin_utama"`
	Password string `json:"password" example:"rahasia123"`
}

type DevicePendingList struct {
	ID          int    `json:"id" example:"1"`
	SiswaID     int    `json:"siswa_id" example:"5"`
	NamaSiswa   string `json:"nama_siswa" example:"Hafizh Junior"`
	DeviceToken string `json:"device_token" example:"hp-hafizh-xyz-123"`
	Status      string `json:"status" example:"pending"`
}

type DeviceActionReq struct {
	DeviceID int `json:"device_id" example:"1"`
}

type ResetPassSiswaReq struct {
	NIS         string `json:"nis" example:"22010101"`
	KataKunci   string `json:"kata_kunci" example:"KucingKu"`
	NewPassword string `json:"new_password" example:"SandiBaru123"`
}

// -- MOCK CRUD SISWA (Bisa ditiru untuk Mapel/Kelas) --
type SiswaData struct {
	ID          int    `json:"id" example:"1"`
	NIS         string `json:"nis" example:"22010101"`
	NamaLengkap string `json:"nama_lengkap" example:"Hafizh Junior"`
	Email       string `json:"email" example:"hafizh@example.com"`
}

// -- MOCK SISWA & GURU LAINNYA --
type RiwayatAbsenSiswa struct {
	Tanggal string `json:"tanggal" example:"2026-04-08"`
	Mapel   string `json:"mapel" example:"Matematika"`
	Status  string `json:"status" example:"hadir"`
}

type JadwalGuru struct {
	ID           int    `json:"id" example:"1"`
	Kelas        string `json:"kelas" example:"XII RPL 1"`
	MataPelajaran string `json:"mata_pelajaran" example:"Pemrograman Web"`
	WaktuMulai   string `json:"waktu_mulai" example:"08:00"`
	WaktuSelesai string `json:"waktu_selesai" example:"10:00"`
}

// -- MOCK TAMBAHAN ADMIN --
type AdminStatsResponse struct {
	TotalAkun     int `json:"total_akun" example:"150"`
	AkunAktif     int `json:"akun_aktif" example:"142"`
	AkunInaktif   int `json:"akun_inaktif" example:"8"`
	PendingDevice int `json:"pending_device" example:"3"`
}

type CreateKelasReq struct {
	NamaKelas string `json:"nama_kelas" example:"XII RPL 1"`
}

type CreateMapelReq struct {
	NamaMapel string `json:"nama_mapel" example:"Pemrograman Web & Perangkat Bergerak"`
}

// -- MOCK TAMBAHAN GURU (RESTORE) --
type RestoreDBReq struct {
	FileID string `json:"file_id" example:"1B2a3C4d5E6f7G8h9I0j"`
}