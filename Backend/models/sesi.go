package models

// SesiResponse adalah format data jadwal yang akan dilihat oleh Guru
type SesiResponse struct {
	ID            int    `json:"id"`
	Kelas         string `json:"kelas"`
	MataPelajaran string `json:"mata_pelajaran"`
	WaktuMulai    string `json:"waktu_mulai"`
	WaktuSelesai  string `json:"waktu_selesai"`
}