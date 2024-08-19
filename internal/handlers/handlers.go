package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
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

func HouseCreateHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var house models.House

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &house); err != nil {

			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		house, err = db.CreateHouse(house)
		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(house)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

	})
}

func FlatCreateHandler(db storage.Database, cache storage.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var flat models.Flat
		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &flat); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		flat, err = db.CreateFlat(flat)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(flat)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cache.DeleteFlatsByHouseId(flat.HouseId, `moderator`)
		if flat.Status == `approved` {
			cache.DeleteFlatsByHouseId(flat.HouseId, `client`)
		}

		if err := db.UpdateAtHouseLastFlatTime(flat.HouseId); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func FlatUpdateHandler(db storage.Database, cache storage.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var flat models.Flat

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &flat); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		flat, err = db.UpdateFlat(flat)

		if flat.Id == -1 {
			w.Header().Set("Retry-After", "3")
			http.Error(w, `This apartment is being moderated by another moderator`, http.StatusUnauthorized)

		}

		if err != nil {

			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(flat)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cache.DeleteFlatsByHouseId(flat.HouseId, `moderator`)
		if flat.Status == `approved` {
			cache.DeleteFlatsByHouseId(flat.HouseId, `client`)
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func GetFlatsInHouseHandler(db storage.Database, cache storage.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parameters := mux.Vars(r)

		houseId, err := strconv.ParseInt(parameters[`id`], 10, 64)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userType, ok := r.Context().Value(`userType`).(string)

		if !ok {
			w.Header().Set("Retry-After", "3")
			http.Error(w, `could not get a user type`, http.StatusInternalServerError)
			return
		}

		jsonFlats, err := cache.GetFlatsByHouseID(houseId, userType)

		if err == nil {
			slog.Info(`Flats gets from cache`, "houseID", houseId, "userType", userType)
			w.Header().Set(`Content-Type`, `application/json`)
			w.WriteHeader(http.StatusOK)
			w.Write(jsonFlats)

			return
		}

		flats, err := db.GetFlatsByHouseID(houseId, userType)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := cache.PutFlatsByHouseID(flats, houseId, userType); err != nil {
			slog.Error("Failed to cache flats", "houseID", houseId, "userType", userType, "error", err)
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(flats); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
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
