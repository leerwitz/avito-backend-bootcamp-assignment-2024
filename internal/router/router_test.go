package router

import (
	"avitoBootcamp/internal/models"
	"avitoBootcamp/internal/storage/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestGetFlatsInHouseHandler(t *testing.T) {

	testCases := []struct {
		houseId        int64
		userType       string
		expectedFlats  []models.Flat
		authorized     bool
		expectCacheHit bool
	}{
		// Тест 1: Пользователь с типом "client", авторизован, данные получены из базы данных
		{
			houseId:  1,
			userType: `client`,
			expectedFlats: []models.Flat{
				{Id: 12, HouseId: 100, Price: 199000, Rooms: 3, Num: 10, Status: "created", ModeratorId: 15},
				{Id: 13, HouseId: 100, Price: 250000, Rooms: 4, Num: 12, Status: "approved", ModeratorId: 7},
			},
			authorized:     true,
			expectCacheHit: false,
		},
		// Тест 2: Пользователь с типом "moderator", авторизован, данные получены из кэша
		{
			houseId:  2,
			userType: `moderator`,
			expectedFlats: []models.Flat{
				{Id: 14, HouseId: 200, Price: 300000, Rooms: 5, Num: 20, Status: "on moderation", ModeratorId: 22},
				{Id: 15, HouseId: 200, Price: 350000, Rooms: 6, Num: 25, Status: "declined", ModeratorId: 33},
			},
			authorized:     true,
			expectCacheHit: true,
		},
		// Тест 3: Пользователь не авторизован, должен вернуться код 401
		{
			houseId:        1,
			userType:       `client`,
			expectedFlats:  []models.Flat{},
			authorized:     false, // Пользователь не авторизован
			expectCacheHit: false,
		},
		// Тест 4: Пользователь с типом "client", авторизован, данные отсутствуют в базе данных и кэше
		{
			houseId:        3,
			userType:       `client`,
			expectedFlats:  []models.Flat{},
			authorized:     true,
			expectCacheHit: false,
		},
		// Тест 5: Пользователь с типом "client", авторизован, данные в кэше устарели и должны быть обновлены
		{
			houseId:  4,
			userType: `client`,
			expectedFlats: []models.Flat{
				{Id: 16, HouseId: 300, Price: 400000, Rooms: 7, Num: 30, Status: "approved", ModeratorId: 44},
				{Id: 17, HouseId: 300, Price: 450000, Rooms: 8, Num: 35, Status: "created", ModeratorId: 55},
			},
			authorized:     true,
			expectCacheHit: false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d: %s", i, tc.userType), func(t *testing.T) {
			mockDB := new(mocks.Database)
			mockCache := new(mocks.Cache)

			if tc.expectCacheHit {
				cachedData, _ := json.Marshal(tc.expectedFlats)
				mockCache.On("GetFlatsByHouseID", tc.houseId, tc.userType).Return(cachedData, nil).Once()
			} else {
				mockCache.On("GetFlatsByHouseID", tc.houseId, tc.userType).Return(nil, redis.Nil).Once()
				mockDB.On("GetFlatsByHouseID", tc.houseId, tc.userType).Return(tc.expectedFlats, nil).Once()
				mockCache.On("PutFlatsByHouseID", tc.expectedFlats, tc.houseId, tc.userType).Return(nil).Once()
			}

			var token string
			if tc.authorized {
				token, _ = PerformLogin(tc.userType)
			}
			req, err := http.NewRequest("GET", fmt.Sprintf("/house/%d", tc.houseId), nil)
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)

			rr := httptest.NewRecorder()
			handler := New(mockDB, mockCache)
			handler.ServeHTTP(rr, req)

			if !tc.authorized && rr.Code == http.StatusUnauthorized {
				t.Log("Received expected unauthorized status (401)")
				return
			}

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.NotEmpty(t, rr.Body.Bytes(), "Response body should not be empty")

			var flats []models.Flat
			err = json.Unmarshal(rr.Body.Bytes(), &flats)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFlats, flats)

			mockDB.AssertExpectations(t)
			mockCache.AssertExpectations(t)

		})
	}
}

func TestFlatCreateHandler(t *testing.T) {

	testCases := []struct {
		inputFlat       models.Flat
		expectedFlat    models.Flat
		authorized      bool
		expectCacheMiss bool
	}{
		// // Тест 1: Создание квартиры с авторизацией, статус "created"
		{
			inputFlat: models.Flat{
				HouseId: 1, Price: 100000, Rooms: 3, Num: 10, Status: "created",
			},
			expectedFlat: models.Flat{
				Id: 1, HouseId: 1, Price: 100000, Rooms: 3, Num: 10, Status: "created",
			},
			authorized:      true,
			expectCacheMiss: true,
		},
		// // Тест 2: Создание квартиры с авторизацией, статус "approved"
		{
			inputFlat: models.Flat{
				HouseId: 2, Price: 150000, Rooms: 4, Num: 12, Status: "approved",
			},
			expectedFlat: models.Flat{
				Id: 2, HouseId: 2, Price: 150000, Rooms: 4, Num: 12, Status: "approved",
			},
			authorized:      true,
			expectCacheMiss: true,
		},
		// Тест 3: Создание квартиры без авторизации, должен вернуться код 401
		{
			inputFlat: models.Flat{
				HouseId: 3, Price: 200000, Rooms: 5, Num: 15, Status: "created",
			},
			expectedFlat:    models.Flat{},
			authorized:      false, // Пользователь не авторизован
			expectCacheMiss: false,
		},
		// // Тест 4: Создание квартиры с авторизацией, статус "declined"
		{
			inputFlat: models.Flat{
				HouseId: 4, Price: 250000, Rooms: 6, Num: 18, Status: "declined",
			},
			expectedFlat: models.Flat{
				Id: 4, HouseId: 4, Price: 250000, Rooms: 6, Num: 18, Status: "declined",
			},
			authorized:      true,
			expectCacheMiss: true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d: %s", i, tc.inputFlat.Status), func(t *testing.T) {
			mockDB := new(mocks.Database)
			mockCache := new(mocks.Cache)

			mockDB.On("CreateFlat", tc.inputFlat).Return(tc.expectedFlat, nil).Once()

			mockCache.On("DeleteFlatsByHouseId", tc.inputFlat.HouseId, "moderator").Once()
			if tc.expectedFlat.Status == "approved" {
				mockCache.On("DeleteFlatsByHouseId", tc.inputFlat.HouseId, "client").Once()
			}

			mockDB.On("UpdateAtHouseLastFlatTime", tc.inputFlat.HouseId).Return(nil).Once()

			var token string
			if tc.authorized {
				token, _ = PerformLogin("moderator")
			}
			reqBody, _ := json.Marshal(tc.inputFlat)
			req, err := http.NewRequest("POST", "/flat/create", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := New(mockDB, mockCache)
			handler.ServeHTTP(rr, req)

			if !tc.authorized && rr.Code == http.StatusUnauthorized {
				t.Log("Received expected unauthorized status (401)")
				return
			}

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.NotEmpty(t, rr.Body.Bytes(), "Response body should not be empty")

			var createdFlat models.Flat
			err = json.Unmarshal(rr.Body.Bytes(), &createdFlat)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedFlat, createdFlat)

			mockDB.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestHouseCreateHandler(t *testing.T) {

	testCases := []struct {
		name          string
		inputHouse    models.House
		expectedHouse models.House
		authorized    bool
		expectedCode  int
	}{
		// // Тест 1: Успешное создание дома
		{
			name: "Successful house creation",
			inputHouse: models.House{
				Address:   "123 Main St",
				Year:      2020,
				Developer: "Test Developer",
			},
			expectedHouse: models.House{
				Id:        1,
				Address:   "123 Main St",
				Year:      2020,
				Developer: "Test Developer",
			},
			authorized:   true,
			expectedCode: http.StatusOK,
		},
		// // Тест 2: Ошибка десериализации JSON
		{
			name:          "JSON Unmarshal error",
			inputHouse:    models.House{},
			expectedHouse: models.House{},
			authorized:    true,
			expectedCode:  http.StatusBadRequest,
		},
		// Тест 3: Ошибка при создании дома в базе данных
		{
			name: "Database error",
			inputHouse: models.House{
				Address:   "456 Elm St",
				Year:      2021,
				Developer: "Another Developer",
			},
			expectedHouse: models.House{},
			authorized:    true,
			expectedCode:  http.StatusInternalServerError,
		},
		// Тест 4: Неавторизованный доступ
		{
			name:          "Unauthorized access",
			inputHouse:    models.House{},
			expectedHouse: models.House{},
			authorized:    false,
			expectedCode:  http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := new(mocks.Database)
			mockCache := new(mocks.Cache)

			if tc.expectedCode == http.StatusOK {
				mockDB.On("CreateHouse", tc.inputHouse).Return(tc.expectedHouse, nil).Once()
			} else if tc.expectedCode == http.StatusInternalServerError {
				mockDB.On("CreateHouse", tc.inputHouse).Return(models.House{}, errors.New("database error")).Once()
			}

			var token string
			if tc.authorized {
				token, _ = PerformLogin("moderator")
			}

			var body []byte
			if tc.expectedCode == http.StatusBadRequest {
				body = []byte(`{"invalidJson"}`) // Неверный JSON
			} else {
				body, _ = json.Marshal(tc.inputHouse)
			}

			req, err := http.NewRequest("POST", "/house/create", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := New(mockDB, mockCache)
			handler.ServeHTTP(rr, req)

			if !tc.authorized {
				assert.Equal(t, http.StatusUnauthorized, rr.Code)
				mockDB.AssertNotCalled(t, "CreateHouse", tc.inputHouse)
				return
			}

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var house models.House
				err = json.Unmarshal(rr.Body.Bytes(), &house)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedHouse, house)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestFlatUpdateHandler(t *testing.T) {

	testCases := []struct {
		name             string
		inputFlat        models.Flat
		updatedFlat      models.Flat
		authorized       bool
		expectedCode     int
		expectCacheClear bool
	}{
		// Тест 1: Успешное обновление квартиры, авторизован
		{
			name: "Successful update by moderator",
			inputFlat: models.Flat{
				Id: 12, HouseId: 100, Price: 199000, Rooms: 3, Num: 10, Status: "created", ModeratorId: 15,
			},
			updatedFlat: models.Flat{
				Id: 12, HouseId: 100, Price: 199000, Rooms: 3, Num: 10, Status: "approved", ModeratorId: 15,
			},
			authorized:       true,
			expectedCode:     http.StatusOK,
			expectCacheClear: true,
		},
		// Тест 2: Ошибка десериализации JSON
		{
			name:             "JSON Unmarshal error",
			inputFlat:        models.Flat{},
			updatedFlat:      models.Flat{},
			authorized:       true,
			expectedCode:     http.StatusBadRequest,
			expectCacheClear: false,
		},
		// Тест 3: Неавторизованный запрос
		{
			name: "Unauthorized access",
			inputFlat: models.Flat{
				Id: 12, HouseId: 100, Price: 199000, Rooms: 3, Num: 10, Status: "created", ModeratorId: 15,
			},
			updatedFlat:  models.Flat{},
			authorized:   false,
			expectedCode: http.StatusUnauthorized,
		},
		// Тест 4: Ошибка обновления базы данных
		{
			name: "Database update error",
			inputFlat: models.Flat{
				Id: 12, HouseId: 100, Price: 199000, Rooms: 3, Num: 10, Status: "created", ModeratorId: 15,
			},
			updatedFlat:  models.Flat{},
			authorized:   true,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test case %d: %s", i, tc.name), func(t *testing.T) {
			mockDB := new(mocks.Database)
			mockCache := new(mocks.Cache)

			if tc.expectedCode == http.StatusOK {
				mockDB.On("UpdateFlat", tc.inputFlat).Return(tc.updatedFlat, nil).Once()
				if tc.expectCacheClear {
					mockCache.On("DeleteFlatsByHouseId", tc.inputFlat.HouseId, "moderator").Return(nil).Once()
					if tc.updatedFlat.Status == "approved" {
						mockCache.On("DeleteFlatsByHouseId", tc.inputFlat.HouseId, "client").Return(nil).Once()
					}
				}
			} else if tc.expectedCode == http.StatusInternalServerError {
				mockDB.On("UpdateFlat", tc.inputFlat).Return(models.Flat{}, errors.New("database error")).Once()
			}

			var token string
			if tc.authorized {
				token, _ = PerformLogin("moderator")
			}

			var body []byte
			if tc.expectedCode == http.StatusBadRequest {
				body = []byte(`{"invalidJson"}`) // Неверный JSON
			} else {
				body, _ = json.Marshal(tc.inputFlat)
			}

			req, err := http.NewRequest("POST", "/flat/update", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Authorization", token)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := New(mockDB, mockCache)
			handler.ServeHTTP(rr, req)

			if !tc.authorized && rr.Code == http.StatusUnauthorized {
				t.Log("Received expected unauthorized status (401)")
				return
			}

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var flat models.Flat
				err = json.Unmarshal(rr.Body.Bytes(), &flat)
				assert.NoError(t, err)
				assert.Equal(t, tc.updatedFlat, flat)
			}

			mockDB.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
