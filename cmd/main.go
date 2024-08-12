package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {

	driverName := "postgres"
	databaseName := "user=postgres password=1980 dbname=postgres host=10.0.2.15 port=5432 sslmode=disable"

	database, err := sql.Open(driverName, databaseName)

	if err != nil {
		log.Fatal(err)
	}

	if err := database.Ping(); err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	router := mux.NewRouter()

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{`*`},
		AllowedMethods:   []string{`GET`, `POST`, `DELETE`, `OPTIONS`, `PATCH`, `PUT`},
		AllowedHeaders:   []string{`Content-Type", "Authorization`},
		AllowCredentials: true,
	}).Handler(router)

	log.Fatal(http.ListenAndServe(`:8080`, handler))

}
