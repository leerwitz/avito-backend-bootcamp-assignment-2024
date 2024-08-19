package handlers

import (
	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func AuthorizationMiddleware(next http.Handler, onlyModerator bool, db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Retry-After", "3")
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &models.CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims.UserId != `dummyLogin` {
			_, err := db.GetUserById(claims.UserId)
			if err != nil {
				w.Header().Set("Retry-After", "3")
				http.Error(w, `Invalid authorization token`, http.StatusUnauthorized)
				return
			}
		}

		if onlyModerator && claims.Type != `moderator` {
			w.Header().Set("Retry-After", "3")
			http.Error(w, "You are not a moderator", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), `userType`, claims.Type)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
