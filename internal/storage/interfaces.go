package storage

import (
	"avitoBootcamp/internal/models"
)

type Database interface {
	GetFlatsByHouseID(houseId int64, userType string) ([]models.Flat, error)
	CreateFlat(flat models.Flat) (models.Flat, error)
	UpdateAtHouseLastFlatTime(houseId int64) error
	CreateHouse(house models.House) (models.House, error)
	UpdateFlat(flat models.Flat) (models.Flat, error)
}

type Cache interface {
	PutFlatsByHouseID(flats []models.Flat, houseId int64, userType string) error
}
