package models

// SiswaRequest adalah data yang dikirim Admin dari Frontend/Postman
type SiswaRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	NamaSekolah    string `json:"nama_sekolah"`
	NamaLengkap    string `json:"nama_lengkap"`
	LabelKataKunci string `json:"label_kata_kunci"` // Clue keamanan yang kamu minta
	KataKunci      string `json:"kata_kunci"`       // Jawaban clue
}

// SiswaSelfRegisterRequest — payload publik daftar mandiri (device_token = binding perangkat saat daftar)
type SiswaSelfRegisterRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	NamaSekolah    string `json:"nama_sekolah"`
	NamaLengkap    string `json:"nama_lengkap"`
	LabelKataKunci string `json:"label_kata_kunci"`
	KataKunci      string `json:"kata_kunci"`
	DeviceToken    string `json:"device_token"`
}

// SiswaResponse adalah format data aman yang akan kita kirimkan ke layar Admin
type SiswaResponse struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	NamaSekolah    string `json:"nama_sekolah"`
	NamaLengkap    string `json:"nama_lengkap"`
	LabelKataKunci string `json:"label_kata_kunci"`
	// Perhatikan: Kita TIDAK memasukkan password atau kata_kunci di sini demi keamanan!
}

// PendingSiswaRegistration — daftar menunggu persetujuan admin (users.is_active = 0)
type PendingSiswaRegistration struct {
	UserID         int    `json:"user_id"`
	Username       string `json:"username"`
	NamaLengkap    string `json:"nama_lengkap"`
	NamaSekolah    string `json:"nama_sekolah"`
	LabelKataKunci string `json:"label_kata_kunci"`
	CreatedAt      string `json:"created_at"`
}