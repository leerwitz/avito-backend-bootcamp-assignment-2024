package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const jwtKey string = `B2iDZ6286IOLg8O1/f81Zdzh1BglfKTdLVw6twOqZGs=`

type AuthorizationToken struct {
	Token string `json:"token"`
}

type CustomClaims struct {
	Type string
	jwt.RegisteredClaims
}

func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get(`user_type`)

	if userType != `client` && userType != `moderator` {
		http.Error(w, "No such user type", http.StatusInternalServerError)
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &CustomClaims{
		Type: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
	}

	w.Header().Set("Content-type", "application/json")

	json.NewEncoder(w).Encode(AuthorizationToken{Token: tokenStr})
}

func AuthorizationMiddleware(next http.Handler, onlyModerator bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get(`Authorization`)

		claims := &CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "User unauthorized", http.StatusUnauthorized)
		}

		if onlyModerator && claims.Type != `moderator` {
			http.Error(w, "You are not a moderator", http.StatusUnauthorized)
		}

		next.ServeHTTP(w, r)
	})
}
