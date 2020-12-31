package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

var JwtKey = []byte("SecretKey")

func Authenticate(tknStr string) (*UserClaims, error) {
	claims := &UserClaims{}

	// Parse the JWT string and store the result in `claims`.
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}
	if !tkn.Valid {
		return nil, errors.New("Not valid")
	}

	return claims, nil
}
