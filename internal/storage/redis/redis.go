package redis

import (
	"avitoBootcamp/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

func NewForTest() (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
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
		slog.Error("Failed to marshal flats", slog.Any("err", err))
		return err
	}

	keyRequest := fmt.Sprintf(`houseID:%d,userType:%s`, houseId, userType)
	request := r.Client.Set(ctx, keyRequest, jsonFlats, 5*time.Minute)

	if err := request.Err(); err != nil {
		slog.Error("Failed to set flats in cache", slog.Any("err", err))
		return err
	}

	slog.Info("Successfully cached flats", "key", keyRequest)

	return nil
}

func (r *RedisCache) GetFlatsByHouseID(houseId int64, userType string) ([]byte, error) {
	ctx := context.Background()
	keyRequest := fmt.Sprintf(`houseID:%d,userType:%s`, houseId, userType)
	request := r.Client.Get(ctx, keyRequest)

	if err := request.Err(); err != nil {
		slog.Error("Failed to get request from the cache", slog.Any("err", err))
		return nil, err
	}

	data, err := request.Result()
	if err != nil {
		slog.Error("Failed to get result from the cache request", slog.Any("err", err))
		return nil, err
	}

	slog.Info(`Successfully get flats from cache`, `key`, keyRequest)

	return []byte(data), nil
}

func (r *RedisCache) DeleteFlatsByHouseId(houseId int64, userType string) {
	ctx := context.Background()
	key := fmt.Sprintf(`houseID:%d,userType:%s`, houseId, userType)

	if err := r.Client.Del(ctx, key).Err(); err != nil {
		slog.Info("Error deleting key:", slog.Any("err", err))
	} else {
		slog.Info("Key deleting successfully")
	}
}
