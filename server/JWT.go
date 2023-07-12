package server

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

/* var api_key = "PortalGun" */
var SECRET = []byte("wubba-lubba-dub-dub")

func GenerateJWT(playerID string) (string, error) {
	/* Generates a JWT Token with one hour expiration time. */

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(time.Hour).Unix()
	claims["PlayerID"] = playerID

	tokenStr, err := token.SignedString(SECRET)

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return tokenStr, nil
}

func extractJWTToken(tokenString string) *jwt.Token {
	/* Extracts token data from JWT Token */

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			println("not authorized: token method is not okay")
		}
		return SECRET, nil
	})
	if err != nil {
		println("not authorized, error: ", err.Error())
	}
	return token
}

func ValidateJWT(tokenString string) bool {
	/* Checks the JWT Token if its stil valid. */

	return extractJWTToken(tokenString).Valid
}

func GetPlayerIDFromJWTToken(tokenString string) string {
	/* Extracts PlayerID from JWT Token payload */

	token := extractJWTToken(tokenString)

	if token.Valid {
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return ""
		}

		PlayerID := claims["PlayerID"].(string)
		return PlayerID

	} else {
		return ""
	}
}
