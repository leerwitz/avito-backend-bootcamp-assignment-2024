package redis

import (
	"avitoBootcamp/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func New() (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisCache{Client: client}, nil
}

func (r *RedisCache) PutFlatsByHouseID(flats []models.Flat, houseId int64, userType string) error {
	ctx := context.Background()

	jsonFlats, err := json.Marshal(flats)

	if err != nil {
		return err
	}

	keyRequest := fmt.Sprintf(`houseID: %d userType: %s`, houseId, userType)
	request := r.Client.Set(ctx, keyRequest, jsonFlats, 5*time.Minute)

	if err := request.Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisCache) GetFlatsByHouseID(flats []models.Flat, houseId int64, userType string) ([]byte, error) {
	ctx := context.Background()
	keyRequest := fmt.Sprintf(`houseID: %d userType: %s`, houseId, userType)
	request := r.Client.Get(ctx, keyRequest)

	if err := request.Err(); err != nil {
		return nil, err
	}

	data, _ := request.Result()

	return []byte(data), nil
}
