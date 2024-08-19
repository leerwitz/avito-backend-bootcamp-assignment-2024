package storage

import (
	"avitoBootcamp/internal/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=database
type Database interface {
	GetFlatsByHouseID(houseId int64, userType string) ([]models.Flat, error)
	CreateFlat(flat models.Flat) (models.Flat, error)
	UpdateAtHouseLastFlatTime(houseId int64) error
	CreateHouse(house models.House) (models.House, error)
	UpdateFlat(flat models.Flat) (models.Flat, error)
	CreateUser(user models.User) (models.User, error)
	GetUserById(id string) (models.User, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=cache
type Cache interface {
	PutFlatsByHouseID(flats []models.Flat, houseId int64, userType string) error
	GetFlatsByHouseID(houseId int64, userType string) ([]byte, error)
	DeleteFlatsByHouseId(houseId int64, userType string)
}
