package auth

import (
	"errors"
	"net/http"
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

func TestGetBearer(t *testing.T) {
	cases := []struct {
		name   string
		bearer string
		header http.Header
		err    error
	}{
		{
			name:   "valid",
			bearer: "thisIsATokenString",
			header: map[string][]string{
				"Authorization": {"Bearer thisIsATokenString"},
			},
		},
		{
			name:   "valid with whitespace",
			bearer: "thisIsATokenString",
			header: map[string][]string{
				"Authorization": {"Bearer   thisIsATokenString   "},
			},
		},
		{
			name:   "invalid auth header - missing",
			bearer: "thisIsATokenString",
			header: map[string][]string{},
			err:    errors.New("invalid authorization header format"),
		},
		{
			name:   "invalid auth header - typo",
			bearer: "thisIsATokenString",
			header: map[string][]string{
				"Oopsorization": {"Bearer thisIsATokenString"},
			},
			err: errors.New("invalid authorization header format"),
		},
		{
			name:   "invalid auth header - no bearer prefix",
			bearer: "thisIsATokenString",
			header: map[string][]string{
				"Authorization": {"somethingElse thisIsATokenString"},
			},
			err: errors.New("invalid authorization header format"),
		},
	}

	for _, c := range cases {
		bearer, err := GetBearerToken(c.header)
		if c.err == nil && err != nil {
			t.Errorf("case '%s': unable to get bearer token: %s", c.name, err)
		}

		if c.err != nil {
			if c.err.Error() != err.Error() {
				t.Errorf("case '%s': unexpected error, expected: '%s', got: '%s'", c.name, c.err, err)
			}
		} else if bearer != c.bearer {
			t.Errorf("case '%s': expected: '%s', got: '%s'", c.name, c.bearer, bearer)
		}
	}
}
