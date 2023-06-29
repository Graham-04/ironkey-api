package hash

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func GeneratePasswordHash(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		log.Fatal("[hash.go] Could not generate password hash. Exiting...")
	}
	return string(hashedPassword)
}
