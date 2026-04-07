package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	
)

// InitDriveService menghubungkan ke Google Drive menggunakan OAuth 2.0
func InitDriveService() (*drive.Service, error) {
	ctx := context.Background()

	// 1. Membaca file client_secret.json
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		return nil, fmt.Errorf("gagal membaca client_secret.json: %v", err)
	}

	// 2. Meminta akses penuh ke Drive
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("gagal memparsing config: %v", err)
	}

	// 3. Mendapatkan token (baca dari file atau minta login di web)
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		// Jika token belum ada, minta user login lewat browser
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	client := config.Client(ctx, tok)

	// 4. Membuat Drive Service
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat layanan Drive: %v", err)
	}

	return srv, nil
}

// UploadToDrive mengunggah file ke folder tertentu
func UploadToDrive(srv *drive.Service, fileName string, folderID string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	driveFile := &drive.File{
		Name:    fileName,
		Parents: []string{folderID}, // Folder tempat file disimpan
	}

	_, err = srv.Files.Create(driveFile).Media(f).Do()
	return err
}

// DownloadFromDrive mengunduh file dari Drive berdasarkan fileID
func DownloadFromDrive(srv *drive.Service, fileID string, localPath string) error {
	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ==========================================
// FUNGSI INTERNAL UNTUK MENGURUS TOKEN OAUTH
// ==========================================

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n=== LOGIN GOOGLE DIPERLUKAN ===\n")
	fmt.Printf("Buka link ini di browsermu:\n\n%v\n\n", authURL)
	fmt.Printf("Setelah login, copy kodenya dan paste di sini, lalu tekan Enter: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Gagal membaca kode otorisasi: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Gagal menukar kode dengan token: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Gagal menyimpan token ke file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}