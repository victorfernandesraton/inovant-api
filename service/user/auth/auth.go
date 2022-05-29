package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordGen generates the password hash
func PasswordGen(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), 12)
}
