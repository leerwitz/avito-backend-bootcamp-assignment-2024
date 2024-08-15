package main

import (
	"avitoBootcamp/internal/handlers"
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	driverName := "postgres"
	databaseName := "user=postgres password=postgres dbname=avitobootcamp host=10.0.2.15 port=5432 sslmode=disable"

	database, err := sql.Open(driverName, databaseName)

	if err != nil {
		log.Fatal(err)
	}

	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	log.Println("Successfully connected to the database!")

	router := mux.NewRouter()

	router.HandleFunc(`/dummyLogin`, handlers.DummyLoginHandler).Methods(`GET`)
	router.Handle(`/house/{id}`, handlers.AuthorizationMiddleware(handlers.GetFlatsInHouseHandler(database), false)).Methods(`GET`)
	router.Handle(`/flat/create`, handlers.AuthorizationMiddleware(handlers.FlatCreateHandler(database), false)).Methods(`POST`)
	router.Handle(`/house/create`, handlers.AuthorizationMiddleware(handlers.HouseCreateHandler(database), true)).Methods(`POST`)
	router.Handle(`/flat/update`, handlers.AuthorizationMiddleware(handlers.FlatUpdateHandler(database), true)).Methods(`POST`)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{`*`},
		AllowedMethods:   []string{`GET`, `POST`, `DELETE`, `OPTIONS`, `PATCH`, `PUT`},
		AllowedHeaders:   []string{`Content-Type", "Authorization`},
		AllowCredentials: true,
	}).Handler(router)

	log.Fatal(http.ListenAndServe(`:8080`, handler))

}
