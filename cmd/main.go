package main

import (
	"avitoBootcamp/internal/router"
	"avitoBootcamp/internal/storage/postgres"
	"avitoBootcamp/internal/storage/redis"
	"log"
	"log/slog"
	"net/http"
)

func main() {
	database, err := postgres.New()

	if err != nil {
		log.Fatal(err)
	}

	defer database.Db.Close()

	slog.Info("Successfully connected to the database!")

	redisClient, err := redis.New()

	if err != nil {
		log.Fatal(err)
	}

	slog.Info(`Successfully connected to the redis client!`)

	handler := router.New(database, redisClient)

	log.Fatal(http.ListenAndServe(`:8080`, handler))

}
