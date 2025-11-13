package models

import "github.com/google/uuid"

type Subscription struct {
	ID          int       `json:"id" 				gorm:"primary_key" 			db:"id"`
	ServiceName string    `json:"service_name" 		gorm:"size:100; not null" 	db:"service_name"`
	Price       int       `json:"price"				gorm:"not null"				db:"price"`
	UserID      uuid.UUID `json:"user_id"			gorm:"type:uuid; not null"	db:"user_id"`
	StartDate   string    `json:"start_date" 		gorm:"size:7; not null"		db:"start_date"`
	EndDate     string    `json:"end_date" 			gorm:"size:7; not null"		db:"end_date"`
}
