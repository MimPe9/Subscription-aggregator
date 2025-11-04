package models

import "github.com/google/uuid"

type Subscription struct {
	ServiceName string    `json:"service_name" 	db:"service_name"`
	Price       int       `json:"price"			db:"price"`
	UserId      uuid.UUID `json:"user_id"		db:"user_id"`
	StartDate   string    `json:"start_date"	db:"start_date"`
}
