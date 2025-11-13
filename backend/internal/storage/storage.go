package storage

import (
	"subscription-aggregator/backend/internal/models"

	"github.com/google/uuid"
)

type Storage interface {
	Create(sub *models.Subscription) error
	Delete(id int) error
	Close() error
	Update(sub *models.Subscription) error
	GetAllEntries() ([]models.Subscription, error)
	GetOneEntry(id int) (*models.Subscription, error)
	SumPrice(start_date, end_date, ServiceName string, UserID uuid.UUID) (int, error)
}
