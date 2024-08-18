package tests

import (
	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/router"
	"avitoBootcamp/internal/storage/postgres"
	"avitoBootcamp/internal/storage/redis"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFlats(t *testing.T) {
	testCases := []struct {
		name            string
		houseId         int64
		userType        string
		expectedFlats   []models.Flat
		authorized      bool
		expectedCode    int
		expectCacheHit  bool
		expectCacheData bool
	}{

		// Тест 1: Пользователь с типом "client", авторизован, данные получены из базы данных
		{
			name:     "Authorized client, data from DB",
			houseId:  1,
			userType: "client",
			expectedFlats: []models.Flat{
				{Id: 2, HouseId: 1, Price: 150000, Rooms: 4, Num: 102, Status: "approved", ModeratorId: 1},
			},
			authorized:      true,
			expectedCode:    http.StatusOK,
			expectCacheHit:  false,
			expectCacheData: true,
		},
		// Тест 2: Пользователь не авторизован, должен вернуться код 401
		{
			name:            "Unauthorized access",
			houseId:         1,
			userType:        "client",
			expectedFlats:   nil,
			authorized:      false,
			expectedCode:    http.StatusUnauthorized,
			expectCacheHit:  false,
			expectCacheData: false,
		},
		// Тест 3: Пользователь с типом "client", авторизован, данные уже есть в кэше
		{
			name:     "Authorized client, data from cache",
			houseId:  1,
			userType: "client",
			expectedFlats: []models.Flat{
				{Id: 2, HouseId: 1, Price: 150000, Rooms: 4, Num: 102, Status: "approved", ModeratorId: 1},
			},
			authorized:      true,
			expectedCode:    http.StatusOK,
			expectCacheHit:  false,
			expectCacheData: false,
		},
		// Тест 4: Запрос с несуществующим houseId
		{
			name:            "Non-existing houseId",
			houseId:         9999,
			userType:        "client",
			expectedFlats:   nil,
			authorized:      true,
			expectedCode:    http.StatusOK,
			expectCacheHit:  false,
			expectCacheData: false,
		},
		// Тест 5: Пользователь с типом "moderator", авторизован, данные получены из базы данных
		{
			name:     "Authorized moderator, data from DB",
			houseId:  1,
			userType: "moderator",
			expectedFlats: []models.Flat{
				{Id: 1, HouseId: 1, Price: 100000, Rooms: 3, Num: 101, Status: "created", ModeratorId: 0},
				{Id: 2, HouseId: 1, Price: 150000, Rooms: 4, Num: 102, Status: "approved", ModeratorId: 1},
			},
			authorized:      true,
			expectedCode:    http.StatusOK,
			expectCacheHit:  false,
			expectCacheData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := postgres.ConnectForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к базе данных: %v", err)
			}
			defer db.Db.Close()

			cache, err := redis.NewForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к клиенту redis: %v", err)
			}
			defer cache.Client.Close()

			// Заранее сохраняем данные в кеш для теста с кэшированием
			if tc.expectCacheHit {
				cachedData, _ := json.Marshal(tc.expectedFlats)
				cacheKey := fmt.Sprintf("houseID:%d,userType:%s", tc.houseId, tc.userType)
				cache.Client.Set(context.Background(), cacheKey, cachedData, 0)
			}

			// Получаем токен, если пользователь авторизован
			var token string
			if tc.authorized {
				token, err = router.PerformLogin(tc.userType)
				if err != nil {
					t.Fatalf("Не удалось получить токен: %v", err)
				}
			}

			req, err := http.NewRequest("GET", fmt.Sprintf("/house/%d", tc.houseId), nil)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			rr := httptest.NewRecorder()
			handler := router.New(db, cache)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			// Если запрос успешен, проверяем тело ответа
			if tc.expectedCode == http.StatusOK {
				assert.NotEmpty(t, rr.Body.Bytes(), "Response body should not be empty")

				var flats []models.Flat
				err = json.Unmarshal(rr.Body.Bytes(), &flats)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFlats, flats)
			}

			// Проверка кеша
			if tc.expectCacheData {
				cacheKey := fmt.Sprintf("houseID:%d,userType:%s", tc.houseId, tc.userType)
				cachedData, err := cache.Client.Get(context.Background(), cacheKey).Result()
				assert.NoError(t, err)
				assert.NotEmpty(t, cachedData, "Cached data should not be empty")

				var cachedFlats []models.Flat
				err = json.Unmarshal([]byte(cachedData), &cachedFlats)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFlats, cachedFlats)
			}
		})
	}
}

func TestFlatCreateHandler(t *testing.T) {
	testCases := []struct {
		name             string
		inputFlat        models.Flat
		userType         string
		authorized       bool
		expectedCode     int
		expectCacheClear bool
	}{
		// Тест 1: Успешное создание квартиры, авторизованный пользователь
		{
			name: "Authorized user, successful creation",
			inputFlat: models.Flat{
				HouseId: 1, Price: 200000, Rooms: 3, Num: 103, Status: "created", ModeratorId: 1,
			},
			userType:         "moderator",
			authorized:       true,
			expectedCode:     http.StatusOK,
			expectCacheClear: true,
		},
		// Тест 2: Неавторизованный запрос
		{
			name: "Unauthorized access",
			inputFlat: models.Flat{
				HouseId: 1, Price: 200000, Rooms: 3, Num: 103, Status: "created", ModeratorId: 1,
			},
			userType:         "client",
			authorized:       false,
			expectedCode:     http.StatusUnauthorized,
			expectCacheClear: false,
		},
		// Тест 3: Ошибка валидации входных данных (некорректный JSON)
		{
			name:             "Invalid JSON",
			inputFlat:        models.Flat{},
			userType:         "moderator",
			authorized:       true,
			expectedCode:     http.StatusBadRequest,
			expectCacheClear: false,
		},
		// Тест 4: Ошибка базы данных при создании квартиры
		{
			name: "Database error",
			inputFlat: models.Flat{
				HouseId: 9999, Price: 200000, Rooms: 3, Num: 103, Status: "created", ModeratorId: 1,
			},
			userType:         "moderator",
			authorized:       true,
			expectedCode:     http.StatusInternalServerError,
			expectCacheClear: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := postgres.ConnectForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к базе данных: %v", err)
			}
			defer db.Db.Close()

			cache, err := redis.NewForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к клиенту redis: %v", err)
			}
			defer cache.Client.Close()

			var token string
			if tc.authorized {
				token, err = router.PerformLogin(tc.userType)
				if err != nil {
					t.Fatalf("Не удалось получить токен: %v", err)
				}
			}

			// Сериализация структуры inputFlat в JSON
			body, err := json.Marshal(tc.inputFlat)
			if tc.name == "Invalid JSON" {
				body = []byte(`{"invalid_json"`) // Некорректный JSON
			}
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/flat/create", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := router.New(db, cache)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var createdFlat models.Flat
				err = json.Unmarshal(rr.Body.Bytes(), &createdFlat)
				assert.NoError(t, err)
				assert.Equal(t, tc.inputFlat.HouseId, createdFlat.HouseId)
				assert.Equal(t, tc.inputFlat.Price, createdFlat.Price)
				assert.Equal(t, tc.inputFlat.Rooms, createdFlat.Rooms)
				assert.Equal(t, tc.inputFlat.Num, createdFlat.Num)
			}

			// Проверка, что кеш был очищен при успешном создании
			if tc.expectCacheClear {
				cacheKeyModerator := fmt.Sprintf("houseID:%d,userType:moderator", tc.inputFlat.HouseId)
				cachedData, err := cache.Client.Get(context.Background(), cacheKeyModerator).Result()
				assert.Error(t, err) // Ошибка должна быть, так как ключ должен быть удален
				assert.Empty(t, cachedData)

				if tc.inputFlat.Status == "approved" {
					cacheKeyClient := fmt.Sprintf("houseID:%d,userType:client", tc.inputFlat.HouseId)
					cachedData, err = cache.Client.Get(context.Background(), cacheKeyClient).Result()
					assert.Error(t, err) // Ошибка должна быть, так как ключ должен быть удален
					assert.Empty(t, cachedData)
				}
			}
		})
	}
}

func TestHouseCreateHandler(t *testing.T) {
	testCases := []struct {
		name         string
		inputHouse   models.House
		userType     string
		authorized   bool
		expectedCode int
	}{
		// Тест 1: Успешное создание дома, авторизованный пользователь
		{
			name: "Authorized user, successful house creation",
			inputHouse: models.House{
				Address:   "123 New Street",
				Year:      2021,
				Developer: "New Developer Inc.",
				CreatedAt: "2024-08-18T12:00:00.000Z",
				UpdateAt:  "2024-08-18T12:00:00.000Z",
			},
			userType:     "moderator",
			authorized:   true,
			expectedCode: http.StatusOK,
		},
		// Тест 2: Неавторизованный запрос
		{
			name: "Unauthorized access",
			inputHouse: models.House{
				Address:   "123 New Street",
				Year:      2021,
				Developer: "New Developer Inc.",
				CreatedAt: "2024-08-18T12:00:00.000Z",
				UpdateAt:  "2024-08-18T12:00:00.000Z",
			},
			userType:     "client",
			authorized:   false,
			expectedCode: http.StatusUnauthorized,
		},
		// Тест 3: Ошибка валидации входных данных (некорректный JSON)
		{
			name:         "Invalid JSON",
			inputHouse:   models.House{},
			userType:     "moderator",
			authorized:   true,
			expectedCode: http.StatusBadRequest,
		},
		// Тест 4: Ошибка базы данных при создании дома
		{
			name: "Database error",
			inputHouse: models.House{
				Address:   "456 Error Street",
				Year:      -22,
				Developer: "Error Developer Inc.",
				CreatedAt: "2024-08-18T12:00:00.000Z",
				UpdateAt:  "2024-08-18T12:00:00.000Z",
			},
			userType:     "moderator",
			authorized:   true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := postgres.ConnectForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к базе данных: %v", err)
			}
			defer db.Db.Close()

			// Получаем токен, если пользователь авторизован
			var token string
			if tc.authorized {
				token, err = router.PerformLogin(tc.userType)
				if err != nil {
					t.Fatalf("Не удалось получить токен: %v", err)
				}
			}

			// Подготовка тела запроса
			body, err := json.Marshal(tc.inputHouse)
			if tc.name == "Invalid JSON" {
				body = []byte(`{"invalid_json"`) // Некорректный JSON
			}
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/house/create", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := router.New(db, nil)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			// Если запрос успешен, проверяем тело ответа
			if tc.expectedCode == http.StatusOK {
				var createdHouse models.House
				err = json.Unmarshal(rr.Body.Bytes(), &createdHouse)
				assert.NoError(t, err)
				assert.Equal(t, tc.inputHouse.Address, createdHouse.Address)
				assert.Equal(t, tc.inputHouse.Year, createdHouse.Year)
				assert.Equal(t, tc.inputHouse.Developer, createdHouse.Developer)
			}
		})
	}
}

func TestFlatUpdateHandler(t *testing.T) {
	testCases := []struct {
		name             string
		inputFlat        models.Flat
		userType         string
		authorized       bool
		expectedCode     int
		expectedFlat     models.Flat
		expectCacheClear bool
	}{
		// Тест 1: Успешное обновление квартиры, авторизованный модератор
		{
			name: "Authorized moderator, successful flat update",
			inputFlat: models.Flat{
				Id: 1, Status: "approved", ModeratorId: 3,
			},
			userType:     "moderator",
			authorized:   true,
			expectedCode: http.StatusOK,
			expectedFlat: models.Flat{
				Id: 1, HouseId: 1, Price: 100000, Rooms: 3, Num: 101, Status: "approved", ModeratorId: 0,
			},
			expectCacheClear: true,
		},
		// Тест 2: Неавторизованный запрос
		{
			name: "Unauthorized access",
			inputFlat: models.Flat{
				Id: 1, Status: "approved", ModeratorId: 21,
			},
			userType:         "client",
			authorized:       true,
			expectedCode:     http.StatusUnauthorized,
			expectedFlat:     models.Flat{},
			expectCacheClear: false,
		},
		// Тест 3: Ошибка при некорректном JSON
		{
			name: "Invalid JSON",
			inputFlat: models.Flat{
				Id: 1, Status: "approved", ModeratorId: 10,
			},
			userType:         "moderator",
			authorized:       true,
			expectedCode:     http.StatusBadRequest,
			expectedFlat:     models.Flat{},
			expectCacheClear: false,
		},
		// Тест 4: Ошибка базы данных при обновлении квартиры
		{
			name: "Database update error",
			inputFlat: models.Flat{
				Id: 9999, Status: "approved", ModeratorId: 0,
			},
			userType:         "moderator",
			authorized:       true,
			expectedCode:     http.StatusInternalServerError,
			expectedFlat:     models.Flat{},
			expectCacheClear: false,
		},
		// Тест 5: Успешная смена статуса квартиры на "on moderation"
		{
			name: "Changing status to “on moderation”",
			inputFlat: models.Flat{
				Id: 1, Status: "on moderation", ModeratorId: 13,
			},
			userType:     "moderator",
			authorized:   true,
			expectedCode: http.StatusOK,
			expectedFlat: models.Flat{
				Id: 1, HouseId: 1, Price: 100000, Rooms: 3, Num: 101, Status: "on moderation", ModeratorId: 13,
			},
			expectCacheClear: false,
		},
		// Тест 6: Ошибка изменить статус квартиры на модерации другим модератором
		{
			name: "Changing status to “on moderation”",
			inputFlat: models.Flat{
				Id: 1, Status: "declined", ModeratorId: 47,
			},
			userType:         "moderator",
			authorized:       true,
			expectedCode:     http.StatusUnauthorized,
			expectedFlat:     models.Flat{},
			expectCacheClear: false,
		},
		// Тест 5: Успешная смена статуса квартиры на "declined"
		{
			name: "Changing status to “declined”",
			inputFlat: models.Flat{
				Id: 1, Status: "declined", ModeratorId: 13,
			},
			userType:     "moderator",
			authorized:   true,
			expectedCode: http.StatusOK,
			expectedFlat: models.Flat{
				Id: 1, HouseId: 1, Price: 100000, Rooms: 3, Num: 101, Status: "declined", ModeratorId: 13,
			},
			expectCacheClear: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := postgres.ConnectForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к базе данных: %v", err)
			}
			defer db.Db.Close()

			cache, err := redis.NewForTest()
			if err != nil {
				t.Fatalf("Не удалось подключиться к клиенту redis: %v", err)
			}
			defer cache.Client.Close()

			// Получаем токен, если пользователь авторизован
			var token string
			if tc.authorized {
				token, err = router.PerformLogin(tc.userType)
				if err != nil {
					t.Fatalf("Не удалось получить токен: %v", err)
				}
			}

			// Подготовка тела запроса
			body, err := json.Marshal(tc.inputFlat)
			if tc.name == "Invalid JSON" {
				body = []byte(`{"invalid_json"`) // Некорректный JSON
			}
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/flat/update", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := router.New(db, cache)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			// Если запрос успешен, проверяем тело ответа
			if tc.expectedCode == http.StatusOK {
				var updatedFlat models.Flat
				err = json.Unmarshal(rr.Body.Bytes(), &updatedFlat)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedFlat, updatedFlat)
			}

			// Проверка, что кеш был очищен при успешном создании
			if tc.expectCacheClear {
				cacheKeyModerator := fmt.Sprintf("houseID:%d,userType:moderator", tc.inputFlat.HouseId)
				cachedData, err := cache.Client.Get(context.Background(), cacheKeyModerator).Result()
				assert.Error(t, err) // Ошибка должна быть, так как ключ должен быть удален
				assert.Empty(t, cachedData)

				if tc.inputFlat.Status == "approved" {
					cacheKeyClient := fmt.Sprintf("houseID:%d,userType:client", tc.inputFlat.HouseId)
					cachedData, err = cache.Client.Get(context.Background(), cacheKeyClient).Result()
					assert.Error(t, err) // Ошибка должна быть, так как ключ должен быть удален
					assert.Empty(t, cachedData)
				}
			}
		})
	}
}
