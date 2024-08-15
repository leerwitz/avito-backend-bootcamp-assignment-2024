package queries

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type House struct {
	Id        int64  `json:"id"`
	Address   string `json:"address"`
	Year      int    `json:"year"`
	Developer string `json:"developer"`
	CreatedAt string `json:"created_at"`
	UpdateAt  string `json:"update_at"`
}

func UpdateAtHouse(db *sql.DB, houseId int64) error {
	currTime := time.Now().UTC().Format(`2017-07-21T17:32:28.000Z`)
	query := `UPDATE house SET update_at = $1 WHERE id = $2`
	_, err := db.Exec(query, currTime, houseId)

	return err
}

func Insert(db *sql.DB, value interface{}, table string) error {
	v := reflect.ValueOf(value)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf(`value must be a pointer to a struct`)
	}

	vType := v.Type()
	numField := v.NumField()

	fieldsValues := make([]interface{}, numField)
	fields := make([]string, numField)
	placeholders := make([]string, numField)

	for i, k := 0, 1; i < numField; i++ {
		field := vType.Field(i)

		if field.Name == `Id` {
			k--
			continue
		}

		fields[i] = field.Tag.Get(`json`)
		placeholders[i] = fmt.Sprintf(`$%d`, i+k)
		fieldsValues[i] = v.Field(i).Interface()
	}

	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s) RETURNING id`,
		table, strings.Join(fields, `,`)[1:], strings.Join(placeholders, `,`)[1:])
	var id int64

	if err := db.QueryRow(query, fieldsValues...).Scan(&id); err != nil {
		return err
	}

	idReflect := reflect.ValueOf(id)

	if v.FieldByName(`Id`).Type() != idReflect.Type() {
		return fmt.Errorf(`id type must be a same`)
	}

	v.FieldByName(`Id`).Set(idReflect)

	fmt.Println(query)

	return nil
}
