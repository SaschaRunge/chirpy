package auth

import (
	"strings"
	"testing"
)

func TestAuth(t *testing.T) {
	password := "suuuup3rl33t"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("unexpected error hashing password '%s': %s", password, err)
	}

	if !strings.Contains(hash, "$argon2id$v=19$m=65536,t=1,p=16$") {
		t.Errorf("hash has unexpected parameters")
	}

	success, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("unexpected error checking password '%s': %s", password, err)
	}
	if !success {
		t.Errorf("failed check password")
	}
}
