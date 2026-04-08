package models

// LoginRequest digunakan untuk menangkap data JSON yang dikirim saat proses login.
// Huruf awalnya Kapital agar bisa dipanggil oleh file lain.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User adalah representasi dari tabel 'users' di database MariaDB kita.
type UserDashboardRow struct {
	ID           int    `json:"id"`
	NamaLengkap  string `json:"nama_lengkap"`
	Username     string `json:"username"` // NIS/NIP sesuai ERD
	Role         string `json:"role"`
	LastLogin    string `json:"last_login"`
	DaysInactive int    `json:"days_inactive"`
	Status       string `json:"status"`
}