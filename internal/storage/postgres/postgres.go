package postgres

import (
	"avitoBootcamp/internal/models"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const (
	host       = "10.0.2.15"
	port       = 5432
	user       = "postgres"
	password   = "postgres"
	dbname     = "avitobootcamp"
	sslmode    = "disable"
	driverName = "postgres"
)

type Storage struct {
	Db *sql.DB
}

func New() (*Storage, error) {
	databaseName := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s", user, password, dbname, host, port, sslmode)
	database, err := sql.Open(driverName, databaseName)

	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		return nil, err
	}

	storage := &Storage{Db: database}

	if err := storage.init(); err != nil {
		return storage, err
	}

	return storage, nil
}

func (storage *Storage) init() error {
	initQuery, err := storage.readSqlQuery(`../tables/createTables.sql`)

	if err != nil {
		return err
	}

	if _, err := storage.Db.Query(initQuery); err != nil {
		return err
	}

	// fillQuery, err := storage.readSqlQuery(`../tables/fillTables.sql`)

	// if err != nil {
	// 	return err
	// }

	// if _, err := storage.Db.Query(fillQuery); err != nil {
	// 	return err
	// }

	return nil
}

func (storage *Storage) readSqlQuery(source string) (string, error) {
	file, err := os.Open(source)

	if err != nil {
		return ``, err
	}

	tables, err := io.ReadAll(file)

	if err != nil {
		return ``, err
	}

	return string(tables), nil
}

func (storage *Storage) GetFlatsByHouseID(houseId int64, userType string) ([]models.Flat, error) {
	query := `SELECT id, house_id, price, rooms, status, flat_num FROM flat  WHERE house_id = $1 `

	if userType != `moderator` {
		query += ` AND status = 'approved'`
	}

	rows, err := storage.Db.Query(query, houseId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var flats []models.Flat

	for rows.Next() {
		var currFlat models.Flat

		if err := rows.Scan(&currFlat.Id, &currFlat.HouseId, &currFlat.Price, &currFlat.Rooms, &currFlat.Status, &currFlat.Num); err != nil {
			return nil, err
		}

		flats = append(flats, currFlat)
	}

	return flats, nil
}

func (storage *Storage) CreateFlat(flat models.Flat) (models.Flat, error) {
	flat.Status = `created`

	query := `INSERT INTO flat (house_id, price, rooms, flat_num, status) 
	VALUES($1, $2, $3, $4, $5) RETURNING id`

	if err := storage.Db.QueryRow(query, flat.HouseId, flat.Price, flat.Rooms, flat.Num, flat.Status).Scan(&flat.Id); err != nil {
		return flat, err
	}

	return flat, nil
}

func (storage *Storage) UpdateAtHouseLastFlatTime(houseId int64) error {
	currTime := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	query := `UPDATE house SET update_at = $1 WHERE id = $2`
	_, err := storage.Db.Exec(query, currTime, houseId)

	return err
}

func (storage *Storage) CreateHouse(house models.House) (models.House, error) {
	house.CreatedAt = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	query := `INSERT INTO house (address, year, developer, created_at) 
		VALUES($1, $2, $3, $4) RETURNING id`

	if err := storage.Db.QueryRow(query, house.Address, house.Year, house.Developer, house.CreatedAt).Scan(&house.Id); err != nil {
		return house, err
	}

	return house, nil
}

func (storage *Storage) UpdateFlat(flat models.Flat) (models.Flat, error) {
	var currStatus string
	var currModeratorId *int

	query := `SELECT status, moderator_id FROM flat WHERE id = $1`
	err := storage.Db.QueryRow(query, flat.Id).Scan(&currStatus, &currModeratorId)
	if err != nil {
		return flat, err
	}

	if currStatus == `on moderation` && (currModeratorId == nil || *currModeratorId != flat.ModeratorId) {
		return flat, err
	}

	if flat.Status == `on moderation` {
		query = `UPDATE flat SET status = $1, moderator_id = $2 WHERE id = $3 RETURNING price, rooms, house_id, flat_num`
		err = storage.Db.QueryRow(query, flat.Status, flat.ModeratorId, flat.Id).Scan(&flat.Price, &flat.Rooms, &flat.HouseId, &flat.Num)
	} else {
		query = `UPDATE flat SET status = $1 WHERE id = $2 RETURNING price, rooms, house_id, flat_num, moderator_id`
		err = storage.Db.QueryRow(query, flat.Status, flat.Id).Scan(&flat.Price, &flat.Rooms, &flat.HouseId, &flat.Num, &currModeratorId)
	}

	if err != nil {
		return flat, err
	}

	if currModeratorId != nil {
		flat.ModeratorId = *currModeratorId
	} else {
		flat.ModeratorId = 0

	}

	return flat, nil
}
