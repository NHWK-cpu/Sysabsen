package models

type Kelas struct {
    ID         int    `json:"id"`
    PeriodeID  int    `json:"periode_id"`
    WaliGuruID *int   `json:"wali_guru_id"` // Tambahan baru (bisa null)
    NamaKelas  string `json:"nama_kelas"`
}