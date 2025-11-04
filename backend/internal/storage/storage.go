package storage

import (
	"subscription-aggregator/backend/internal/models"

	"github.com/google/uuid"
)

type Storage interface {
	Create(sub *models.Subscription) error
	Delete(id uuid.UUID) error
	Close() error
}
