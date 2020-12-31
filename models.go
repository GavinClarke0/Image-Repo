package main

import "github.com/dgrijalva/jwt-go"

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Create a struct that will be encoded to a JWT.
// We add jwt.StandardClaims as an embedded type, to provide fields like expiry time
type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type User struct {
	UserName string
	Password string
	Id       string
}

type ImageMeta struct {
	Path       string
	SamplePath string
	View       string
	Type       string
	Owner      string
	Id         string
}

type GetImageBody struct {
	Name string `json:"name"`
}

type GetImagesBody struct {
	UserName string `json:"userName"`
}
