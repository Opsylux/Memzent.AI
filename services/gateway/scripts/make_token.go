package make_token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func MakeToken() {
	secret := "aura-enterprise-secret-2026"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "admin-01",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Use this Bearer Token for your cURL ===")
	fmt.Printf("Bearer %s\n", tokenString)
}
