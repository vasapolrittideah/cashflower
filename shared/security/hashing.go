package security

import "github.com/matthewhartstonge/argon2"

func HashPassword(password string) (string, error) {
	argon := argon2.DefaultConfig()
	encoded, err := argon.HashEncoded([]byte(password))
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	return argon2.VerifyEncoded([]byte(password), []byte(encodedHash))
}
