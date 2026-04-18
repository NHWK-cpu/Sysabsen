package models

type PeriodeBelajar struct {
	ID          int    `json:"id"`
	TahunAjar   string `json:"tahun_ajaran"`
	Semester    string `json:"semester"`
	StatusAktif int    `json:"status_aktif"`
}
