package main

import (
	"fmt"
	
	// PENTING: Ganti "namamodulekamu" dengan nama module yang ada di file go.mod
	// Misalnya: "backend-absensi/helpers"
	"backend-absensi/helpers" 
)

func main() {
	fmt.Println("Mencoba login ke Google Drive...")
	
	_, err := helpers.InitDriveService()
	if err != nil {
		fmt.Println("Gagal:", err)
		return
	}
	
	fmt.Println("SUKSES! File token.json yang baru berhasil dibuat.")
}