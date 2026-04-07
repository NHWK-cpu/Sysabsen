package controllers

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// Fungsi ini HANYA bisa diakses jika sudah lolos dari Middleware
func DashboardAdmin(w http.ResponseWriter, r *http.Request) {
	// Membaca data yang tadi dititipkan oleh satpam (Middleware) ke dalam Context
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	
	// Mengambil role dari dalam tiket
	role := userInfo["role"].(string)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := fmt.Sprintf(`{"message": "Selamat datang di Ruangan VIP!", "role_anda": "%s"}`, role)
	w.Write([]byte(response))
}