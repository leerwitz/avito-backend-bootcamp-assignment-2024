package handlers

import (
	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const jwtKey string = `B2iDZ6286IOLg8O1/f81Zdzh1BglfKTdLVw6twOqZGs=`

func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get(`user_type`)

	if userType != `client` && userType != `moderator` {
		w.Header().Set("Retry-After", "3")
		http.Error(w, "No such user type", http.StatusInternalServerError)
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &models.CustomClaims{
		UserId: `dummyLogin`,
		Type:   userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtKey))
	if err != nil {

		w.Header().Set("Retry-After", "3")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.AuthorizationToken{Token: tokenStr})
}

func RegisterHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		var user models.User
		if err := json.Unmarshal(body, &user); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			w.Header().Set(`Retry-After`, "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user.Password = string(passwordHash)

		if user, err = db.CreateUser(user); err != nil {
			w.Header().Set(`Retry-After`, "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"user_id": user.Id})

	})
}

func LoginHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set(`Retry-After`, "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		var userFromReq models.User
		if err := json.Unmarshal(body, &userFromReq); err != nil {
			w.Header().Set(`Retry-After`, "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := db.GetUserById(userFromReq.Id)
		if err != nil {
			w.Header().Set(`Retry-After`, "3")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userFromReq.Password)); err != nil {
			w.Header().Set(`Retry-After`, "3")
			http.Error(w, "Invalid password", http.StatusBadRequest)
			return
		}

		expirationTime := time.Now().Add(48 * time.Hour)
		claims := &models.CustomClaims{
			UserId: user.Id,
			Type:   user.UserType,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString([]byte(jwtKey))
		if err != nil {

			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.AuthorizationToken{Token: tokenStr})
	})
}
