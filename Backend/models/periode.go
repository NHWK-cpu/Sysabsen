package models

type PeriodeBelajar struct {
	ID          int    `json:"id"`
	TahunAjar string `json:"tahun_ajar"`
	StatusAktif      string `json:"status_aktif"` 
}