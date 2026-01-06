package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a raw password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ComparePassword compares bcrypt hash with raw password
func ComparePassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),     // ✅ HASH FIRST
		[]byte(password), // ✅ RAW PASSWORD SECOND
	)
}
