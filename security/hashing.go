package security

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword menghasilkan bcrypt hash dari password plaintext.
// cost default = 10. Gunakan untuk hashing password user saat registrasi/ganti password.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash memverifikasi password plaintext terhadap bcrypt hash.
// Mengembalikan true jika cocok, false jika tidak.
// Parameter: hash = bcrypt hash, password = plaintext yang di-verify.
// PENTING: fungsi ini TIDAK boleh logging password plaintext ke console/log.
func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
