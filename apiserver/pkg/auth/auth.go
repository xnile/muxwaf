package auth

import (
	"encoding/base32"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

// HashCost sets the cost of bcrypt hashes
// - if this changes hashed passwords would need to be recalculated.
const HashCost = 10

// CheckPassword compares a password hashed with bcrypt.
func CheckPassword(pass, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
}

// HashPassword hashes a password with a random salt using bcrypt.
func HashPassword(pass string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), HashCost)
	return string(hash)
}

func GenHash() string {
	return base32.StdEncoding.EncodeToString(
		securecookie.GenerateRandomKey(32),
	)
}
