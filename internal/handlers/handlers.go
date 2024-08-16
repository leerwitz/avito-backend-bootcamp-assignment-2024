package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const jwtKey string = `B2iDZ6286IOLg8O1/f81Zdzh1BglfKTdLVw6twOqZGs=`

func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	userType := r.URL.Query().Get(`user_type`)

	if userType != `client` && userType != `moderator` {
		http.Error(w, "No such user type", http.StatusInternalServerError)
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &models.CustomClaims{
		Type: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")

	json.NewEncoder(w).Encode(models.AuthorizationToken{Token: tokenStr})
}

func AuthorizationMiddleware(next http.Handler, onlyModerator bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get(`Authorization`)

		claims := &models.CustomClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if onlyModerator && claims.Type != `moderator` {
			http.Error(w, "You are not a moderator", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), `userType`, claims.Type)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func HouseCreateHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var house models.House

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &house); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// house.CreatedAt = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
		// query := `INSERT INTO house (address, year, developer, created_at)
		// VALUES($1, $2, $3, $4) RETURNING id`

		// if err := db.QueryRow(query, house.Address, house.Year, house.Developer, house.CreatedAt).Scan(&house.Id); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		house, err = db.CreateHouse(house)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(house)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

	})
}

func FlatCreateHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var flat models.Flat
		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &flat); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// flat.Status = `created`

		// query := `INSERT INTO flat (house_id, price, rooms, flat_num, status)
		// VALUES($1, $2, $3, $4, $5) RETURNING id`

		// if err := db.QueryRow(query, flat.HouseId, flat.Price, flat.Rooms, flat.Num, flat.Status).Scan(&flat.Id); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		flat, err = db.CreateFlat(flat)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(flat)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.UpdateAtHouseLastFlatTime(flat.HouseId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func FlatUpdateHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var flat models.Flat

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &flat); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// var currStatus string
		// var currModeratorId *int

		// query := `SELECT status, moderator_id FROM flat WHERE id = $1`
		// err = db.QueryRow(query, flat.Id).Scan(&currStatus, &currModeratorId)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// if currStatus == `on moderation` && (currModeratorId == nil || *currModeratorId != flat.ModeratorId) {
		// 	http.Error(w, `This flat has already been moderated by another moderator`, http.StatusUnauthorized)
		// 	return
		// }

		// if flat.Status == `on moderation` {
		// 	query = `UPDATE flat SET status = $1, moderator_id = $2 WHERE id = $3 RETURNING price, rooms, house_id, flat_num`
		// 	err = db.QueryRow(query, flat.Status, flat.ModeratorId, flat.Id).Scan(&flat.Price, &flat.Rooms, &flat.HouseId, &flat.Num)
		// } else {
		// 	query = `UPDATE flat SET status = $1 WHERE id = $2 RETURNING price, rooms, house_id, flat_num, moderator_id`
		// 	err = db.QueryRow(query, flat.Status, flat.Id).Scan(&flat.Price, &flat.Rooms, &flat.HouseId, &flat.Num, &flat.ModeratorId)
		// }

		flat, err = db.UpdateFlat(flat)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(flat)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func GetFlatsInHouseHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parameters := mux.Vars(r)

		houseId, err := strconv.ParseInt(parameters[`id`], 10, 64)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userType, ok := r.Context().Value(`userType`).(string)

		if !ok {
			http.Error(w, `could not get a user type`, http.StatusInternalServerError)
			return
		}

		// query := `SELECT id, house_id, price, rooms, status, flat_num FROM flat  WHERE house_id = $1 `

		// if userType != `moderator` {
		// 	query += ` AND status = 'approved'`
		// }

		// rows, err := db.Query(query, houseId)

		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// defer rows.Close()

		// var flats []models.Flat

		// for rows.Next() {
		// 	var currFlat models.Flat

		// 	if err := rows.Scan(&currFlat.Id, &currFlat.HouseId, &currFlat.Price, &currFlat.Rooms, &currFlat.Status, &currFlat.Num); err != nil {
		// 		http.Error(w, err.Error(), http.StatusInternalServerError)
		// 		return
		// 	}

		// 	flats = append(flats, currFlat)
		// }

		flats, err := db.GetFlatsByHouseID(houseId, userType)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(flats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
