package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func HouseCreateHandler(db storage.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var house models.House

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &house); err != nil {

			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		house, err = db.CreateHouse(house)
		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(house)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

	})
}

func FlatCreateHandler(db storage.Database, cache storage.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var flat models.Flat
		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &flat); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		flat, err = db.CreateFlat(flat)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(flat)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cache.DeleteFlatsByHouseId(flat.HouseId, `moderator`)
		if flat.Status == `approved` {
			cache.DeleteFlatsByHouseId(flat.HouseId, `client`)
		}

		if err := db.UpdateAtHouseLastFlatTime(flat.HouseId); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func FlatUpdateHandler(db storage.Database, cache storage.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var flat models.Flat

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		if err := json.Unmarshal(body, &flat); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		flat, err = db.UpdateFlat(flat)

		if flat.Id == -1 {
			w.Header().Set("Retry-After", "3")
			http.Error(w, `This apartment is being moderated by another moderator`, http.StatusUnauthorized)

		}

		if err != nil {

			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(flat)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cache.DeleteFlatsByHouseId(flat.HouseId, `moderator`)
		if flat.Status == `approved` {
			cache.DeleteFlatsByHouseId(flat.HouseId, `client`)
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func GetFlatsInHouseHandler(db storage.Database, cache storage.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parameters := mux.Vars(r)

		houseId, err := strconv.ParseInt(parameters[`id`], 10, 64)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userType, ok := r.Context().Value(`userType`).(string)

		if !ok {
			w.Header().Set("Retry-After", "3")
			http.Error(w, `could not get a user type`, http.StatusInternalServerError)
			return
		}

		jsonFlats, err := cache.GetFlatsByHouseID(houseId, userType)

		if err == nil {
			slog.Info(`Flats gets from cache`, "houseID", houseId, "userType", userType)
			w.Header().Set(`Content-Type`, `application/json`)
			w.WriteHeader(http.StatusOK)
			w.Write(jsonFlats)

			return
		}

		flats, err := db.GetFlatsByHouseID(houseId, userType)

		if err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := cache.PutFlatsByHouseID(flats, houseId, userType); err != nil {
			slog.Error("Failed to cache flats", "houseID", houseId, "userType", userType, "error", err)
		}

		w.Header().Set(`Content-Type`, `application/json`)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(flats); err != nil {
			w.Header().Set("Retry-After", "3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
