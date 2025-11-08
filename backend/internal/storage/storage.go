package storage

import (
	"subscription-aggregator/backend/internal/models"

	"github.com/google/uuid"
)

type Storage interface {
	Create(sub *models.Subscription) error
	Delete(id uuid.UUID) error
	Close() error
	Update(sub *models.Subscription) error
	GetAllEntries() ([]models.Subscription, error)
	GetOneEntry(id uuid.UUID) (*models.Subscription, error)
	SumPrice(start, end, ServiceName string, UserId uuid.UUID) (int, error)
}
