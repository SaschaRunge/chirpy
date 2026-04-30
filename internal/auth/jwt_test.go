package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	cases := []struct {
		name             string
		creationSecret   string
		validationSecret string
		id               uuid.UUID
		expiresIn        time.Duration
		err              error
	}{
		{
			name:             "valid token",
			id:               uuid.New(),
			creationSecret:   "thisIsMyTokenSecret",
			validationSecret: "thisIsMyTokenSecret",
			expiresIn:        time.Minute * 10,
			err:              nil,
		},
		{
			name:             "expired token",
			id:               uuid.New(),
			creationSecret:   "thisIsMyTokenSecret",
			validationSecret: "thisIsMyTokenSecret",
			expiresIn:        time.Minute * 0,
			err:              jwt.ErrTokenExpired,
		},
		{
			name:             "signature invalid",
			id:               uuid.New(),
			creationSecret:   "thisIsMyTokenSecret",
			validationSecret: "maliciousAttacker",
			expiresIn:        time.Minute * 10,
			err:              jwt.ErrTokenSignatureInvalid,
		},
		{
			name:             "token invalid",
			id:               uuid.New(),
			creationSecret:   "thisIsMyTokenSecret",
			validationSecret: "thisIsMyTokenSecret",
			expiresIn:        time.Minute * 10,
			err:              jwt.ErrTokenMalformed,
		},
	}

	for _, c := range cases {
		token, err := MakeJWT(c.id, c.creationSecret, c.expiresIn)
		if err != nil {
			t.Errorf("case '%s': unexpected error creating jwt: %s", c.name, err)
		}

		if c.name == "token invalid" {
			token = "thisIsNotAToken"
		}

		_, err = ValidateJWT(token, c.validationSecret)
		if !errors.Is(err, c.err) {
			t.Errorf("case '%s': expected err '%s' got '%s'", c.name, c.err, err)
		}
	}
}
