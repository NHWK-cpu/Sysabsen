package models

type Kelas struct {
	ID        int    `json:"id"`
	PeriodeID int    `json:"periode_id"` // Atribut relasi ERD yang terlewat
	NamaKelas string `json:"nama_kelas"`
}