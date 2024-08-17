package router

import (
	"avitoBootcamp/internal/handlers"
	"avitoBootcamp/internal/storage"
	"net/http"

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
