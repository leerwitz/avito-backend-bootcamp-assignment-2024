package main

import (
	"avitoBootcamp/internal/router"
	"avitoBootcamp/internal/storage/postgres"
	"log"
	"net/http"
)

func main() {
	database, err := postgres.New()

	if err != nil {
		log.Fatal(err)
	}

	defer database.Db.Close()

	log.Println("Successfully connected to the database!")

	// router := mux.NewRouter()

	// router.HandleFunc(`/dummyLogin`, handlers.DummyLoginHandler).Methods(`GET`)
	// router.Handle(`/house/{id}`, handlers.AuthorizationMiddleware(handlers.GetFlatsInHouseHandler(database), false)).Methods(`GET`)
	// router.Handle(`/flat/create`, handlers.AuthorizationMiddleware(handlers.FlatCreateHandler(database), false)).Methods(`POST`)
	// router.Handle(`/house/create`, handlers.AuthorizationMiddleware(handlers.HouseCreateHandler(database), true)).Methods(`POST`)
	// router.Handle(`/flat/update`, handlers.AuthorizationMiddleware(handlers.FlatUpdateHandler(database), true)).Methods(`POST`)

	// handler := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{`*`},
	// 	AllowedMethods:   []string{`GET`, `POST`, `DELETE`, `OPTIONS`, `PATCH`, `PUT`},
	// 	AllowedHeaders:   []string{`Content-Type", "Authorization`},
	// 	AllowCredentials: true,
	// }).Handler(router)
	handler := router.New(database)

	log.Fatal(http.ListenAndServe(`:8080`, handler))

}
