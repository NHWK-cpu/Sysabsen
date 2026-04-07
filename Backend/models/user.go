package models

// LoginRequest digunakan untuk menangkap data JSON yang dikirim saat proses login.
// Huruf awalnya Kapital agar bisa dipanggil oleh file lain.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User adalah representasi dari tabel 'users' di database MariaDB kita.
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"` // Tanda "-" memastikan password tidak akan pernah ikut terkirim ke klien demi keamanan
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}