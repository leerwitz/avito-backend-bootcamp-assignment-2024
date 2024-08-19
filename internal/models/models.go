package models

import "github.com/golang-jwt/jwt/v4"

type AuthorizationToken struct {
	Token string `json:"token"`
}

type CustomClaims struct {
	Type   string
	UserId string
	jwt.RegisteredClaims
}

type House struct {
	Id        int64  `json:"id"`
	Address   string `json:"address"`
	Year      int    `json:"year"`
	Developer string `json:"developer"`
	CreatedAt string `json:"created_at"`
	UpdateAt  string `json:"update_at"`
}

type Flat struct {
	Id          int64  `json:"id"`
	HouseId     int64  `json:"house_id"`
	Price       int64  `json:"price"`
	Rooms       int    `json:"rooms"`
	Status      string `json:"status"`
	Num         int    `json:"flat_num"`
	ModeratorId int    `json:"moderator_id"`
}

type User struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	UserType string `json:"user_type"`
}
