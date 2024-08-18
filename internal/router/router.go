package router

import (
	"avitoBootcamp/internal/handlers"
	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func New(database storage.Database, cache storage.Cache) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc(`/dummyLogin`, handlers.DummyLoginHandler).Methods(`GET`)
	router.Handle(`/house/{id}`, handlers.AuthorizationMiddleware(handlers.GetFlatsInHouseHandler(database, cache), false)).Methods(`GET`)
	router.Handle(`/flat/create`, handlers.AuthorizationMiddleware(handlers.FlatCreateHandler(database, cache), false)).Methods(`POST`)
	router.Handle(`/house/create`, handlers.AuthorizationMiddleware(handlers.HouseCreateHandler(database), true)).Methods(`POST`)
	router.Handle(`/flat/update`, handlers.AuthorizationMiddleware(handlers.FlatUpdateHandler(database, cache), true)).Methods(`POST`)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{`*`},
		AllowedMethods:   []string{`GET`, `POST`, `DELETE`, `OPTIONS`, `PATCH`, `PUT`},
		AllowedHeaders:   []string{`Content-Type", "Authorization`},
		AllowCredentials: true,
	}).Handler(router)

	return handler
}

func PerformLogin(userType string) (string, error) {
	loginReq, err := http.NewRequest("GET", "/dummyLogin?user_type="+userType, nil)
	if err != nil {
		return "", err
	}

	loginRR := httptest.NewRecorder()
	loginHandler := http.HandlerFunc(handlers.DummyLoginHandler)
	loginHandler.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", loginRR.Code)
	}

	var tokenResponse models.AuthorizationToken
	err = json.Unmarshal(loginRR.Body.Bytes(), &tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.Token, nil
}
