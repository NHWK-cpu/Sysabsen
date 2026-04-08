package helpers

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendResetEmail merakit dan mengirim email HTML
func SendResetEmail(toEmail string, resetToken string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	frontendURL := os.Getenv("FRONTEND_URL")

	// Link yang akan diklik guru, mengarah ke Svelte
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	// Format Email HTML
	subject := "Subject: Reset Password Akun Guru\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px; color: #333;">
			<h2>Permintaan Reset Password</h2>
			<p>Seseorang telah meminta untuk mereset password akun Guru Anda.</p>
			<p>Jika ini memang Anda, silakan klik tombol di bawah ini untuk membuat password baru:</p>
			<a href="%s" style="display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px;">Reset Password Sekarang</a>
			<p style="margin-top: 20px; font-size: 12px; color: #777;">Link ini hanya berlaku selama 15 menit. Jika Anda tidak meminta reset, abaikan email ini.</p>
		</div>
	`, resetLink)

	msg := []byte(subject + mime + body)

	// Proses Autentikasi dan Pengiriman
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
	
	return err
}