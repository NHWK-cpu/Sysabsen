package models

type SiswaKelas struct {
	ID        int    `json:"id"`
	SiswaID   int    `json:"siswa_id"`
	KelasID   int    `json:"kelas_id"`
	NamaSiswa string `json:"nama_siswa,omitempty"` // Helper untuk Read
	NamaKelas string `json:"nama_kelas,omitempty"` // Helper untuk Read
}