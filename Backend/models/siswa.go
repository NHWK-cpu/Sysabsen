package models

// SiswaRequest adalah data yang dikirim Admin dari Frontend/Postman
type SiswaRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	NIS            string `json:"nis"`
	NamaLengkap    string `json:"nama_lengkap"`
	LabelKataKunci string `json:"label_kata_kunci"` // Clue keamanan yang kamu minta
	KataKunci      string `json:"kata_kunci"`       // Jawaban clue
}

// SiswaResponse adalah format data aman yang akan kita kirimkan ke layar Admin
type SiswaResponse struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	NIS            string `json:"nis"`
	NamaLengkap    string `json:"nama_lengkap"`
	LabelKataKunci string `json:"label_kata_kunci"`
	// Perhatikan: Kita TIDAK memasukkan password atau kata_kunci di sini demi keamanan!
}