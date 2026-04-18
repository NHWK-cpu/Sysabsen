package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"backend-absensi/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware adalah "Satpam" kita
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Cek apakah klien membawa header "Authorization"
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Akses ditolak. Tiket (Token) tidak ditemukan."}`))
			return
		}

		// 2. Ambil tokennya saja (buang kata "Bearer " di depannya)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Validasi keaslian token menggunakan kunci rahasia dari config
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode enkripsi tidak valid")
			}
			return config.JWT_KEY, nil
		})

		// 4. Jika token rusak, kedaluwarsa, atau palsu
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Akses ditolak. Tiket tidak valid atau sudah kedaluwarsa."}`))
			return
		}

		// 5. Jika lolos, ambil data dari dalam tiket (seperti user_id dan role)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Kita simpan datanya di "Context" agar bisa dibaca oleh Controller tujuan
			ctx := context.WithValue(r.Context(), "userInfo", claims)
			// Lanjutkan perjalanan ke ruangan (Controller) yang dituju!
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Akses ditolak. Data di dalam tiket rusak."}`))
			return
		}
	}
}

// AdminOnly adalah satpam lapis kedua khusus untuk rute Admin
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mengambil data tiket dari JWTMiddleware sebelumnya
		userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
		role := userInfo["role"].(string)

		// Jika rolenya BUKAN admin, tendang keluar!
		if role != "admin" && role != "super_admin" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden) // Status 403: Forbidden
			w.Write([]byte(`{"error": "Akses ditolak. Anda bukan Admin!"}`))
			return
		}

		// Jika dia Admin, persilakan masuk ke Controller
		next.ServeHTTP(w, r)
	}
}

// TAMBAHAN: Satpam lapis ketiga KHUSUS Super Admin
func SuperAdminOnly(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
        role := userInfo["role"].(string)

        // HANYA super_admin yang boleh lewat
        if role != "super_admin" {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusForbidden)
            w.Write([]byte(`{"error": "Akses ditolak. Fitur ini hanya untuk Super Admin!"}`))
            return
        }
        next.ServeHTTP(w, r)
    }
}

// SiswaOnly adalah satpam lapis kedua khusus untuk rute Siswa
func SiswaOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
		role := userInfo["role"].(string)

		if role != "siswa" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "Akses ditolak. Anda bukan Siswa!"}`))
			return
		}
		next.ServeHTTP(w, r)
	}
}

// GuruOnly adalah satpam lapis kedua khusus untuk rute Guru
func GuruOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
		role := userInfo["role"].(string)

		if role != "guru" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "Akses ditolak. Anda bukan Guru!"}`))
			return
		}
		next.ServeHTTP(w, r)
	}
}