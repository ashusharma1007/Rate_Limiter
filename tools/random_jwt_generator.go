package main

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user123",
	})
	s, _ := token.SignedString([]byte("somesecret"))
	fmt.Println(s)
}
