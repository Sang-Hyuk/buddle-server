package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Jwt struct {
    SecretKey string `yaml:"secret_key"`
}

func CreateJWT(Id string, secretKey string) (string, error) {
	mySigningKey := []byte(secretKey)

	aToken := jwt.New(jwt.SigningMethodHS256)
	claims := aToken.Claims.(jwt.MapClaims)
	claims["Id"] = Id
	claims["exp"] = time.Now().Add(time.Minute * 20).Unix()

	tk, err := aToken.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}
	return tk, nil
}