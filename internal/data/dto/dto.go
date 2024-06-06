package dto

import "github.com/golang-jwt/jwt/v5"

type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

// UserDTO struct represents a user in the system
type UserDTO struct {
	ID    string
	Email string
}

// AppDTO struct represents an application with a secret key
type AppDTO struct {
	ID     string
	Secret string
}
