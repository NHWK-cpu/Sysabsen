package models

// Struct untuk menangkap data dari Svelte
type CreateGuruRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	NamaLengkap string `json:"nama_lengkap"`
	NIP         string `json:"identifier"` // Di frontend kita pakai nama 'identifier'
	Email       string `json:"email"`
}

