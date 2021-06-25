package authserv

import (
	"golang.org/x/crypto/bcrypt"
)

// Take a string and return a hash of it
func makeHash(pass string) string {

	// no err needed because cost is OK
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	return string(hash)
}

// pass_hash is the hash from database and
// pass is the one sent by user
func passIsCorrect(pass, pass_hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pass_hash), []byte(pass))
	return err == nil
}
