package storage

import (
	"avitoBootcamp/internal/models"
)

//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=Database
type Database interface {
	GetFlatsByHouseID(houseId int64, userType string) ([]models.Flat, error)
	CreateFlat(flat models.Flat) (models.Flat, error)
	UpdateAtHouseLastFlatTime(houseId int64) error
	CreateHouse(house models.House) (models.House, error)
	UpdateFlat(flat models.Flat) (models.Flat, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.44.2 --name=Cache
type Cache interface {
	PutFlatsByHouseID(flats []models.Flat, houseId int64, userType string) error
	GetFlatsByHouseID(houseId int64, userType string) ([]byte, error)
	DeleteFlatsByHouseId(houseId int64, userType string)
}
