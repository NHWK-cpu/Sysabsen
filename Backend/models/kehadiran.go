package models

// AbsenRequest adalah data yang akan dikirim dari HP Siswa saat menekan tombol absen
// AbsenRequest (Versi Baru untuk QR Dinamis)
type AbsenRequest struct {
	QRToken         string `json:"qr_token"` // Hasil scan kamera HP
	StatusKehadiran string `json:"status_kehadiran"`
	// metode_absen dan sesi_id dihapus karena otomatis diatur oleh backend
}

// GuruAbsenRequest adalah data yang dikirim oleh Guru saat mengabsenkan siswa
type GuruAbsenRequest struct {
	SesiID          int    `json:"sesi_id"`
	SiswaID         int    `json:"siswa_id"`
	StatusKehadiran string `json:"status_kehadiran"`
	Tanggal         string `json:"tanggal"` // <-- TAMBAHKAN INI
}

// AbsenRequest adalah data riwayat absen di hari lampau
type RiwayatAbsenResponse struct {
	Tanggal         string `json:"tanggal"`
	MataPelajaran   string `json:"mata_pelajaran"`
	GuruPengajar    string `json:"guru_pengajar"`
	StatusKehadiran string `json:"status_kehadiran"`
}
