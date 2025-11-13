package storage

import (
	"fmt"
	"log"
	"subscription-aggregator/backend/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStorage struct {
	db *gorm.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	log.Printf("Storage: attempting to connect to database")
	log.Printf("Storage: connection string: %s", connStr)

	var db *gorm.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if err != nil {
			log.Printf("Storage: attempt %d - sql.Open failed: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Storage: attempt %d - get sql.DB failed: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = sqlDB.Ping()
		if err == nil {
			log.Printf("Storage: successfully connected to database on attempt %d", i+1)
			if err := runAutoMigrations(db); err != nil {
				return nil, fmt.Errorf("auto migrations failed: %w", err)
			}
			return &PostgresStorage{db: db}, nil
		}
		log.Printf("Storage: attempt %d - database ping failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Printf("Storage: failed to connect to database after 10 attempts: %v", err)
	return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
}

func runAutoMigrations(db *gorm.DB) error {
	log.Printf("Storage: running GORM auto migrations...")

	err := db.AutoMigrate(&models.Subscription{})
	if err != nil {
		return fmt.Errorf("failed to auto migrate subscriptions: %w", err)
	}

	log.Printf("Storage: GORM auto migrations completed successfully")
	return nil
}

func (s *PostgresStorage) Create(sub *models.Subscription) error {
	log.Printf("Storage.Create: creating subscription for service '%s', price %d, user_id %s, start date %s, end date %s",
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)

	result := s.db.Create(sub)
	if result.Error != nil {
		log.Printf("Storage.Create: failed to create subscription for service '%s' - %v", sub.ServiceName, result.Error)
		return fmt.Errorf("failed to create: %w", result.Error)
	}

	log.Printf("Storage.Create: successfully created subscription with ID %d for service '%s'",
		sub.ID, sub.ServiceName)
	return nil
}

func (s *PostgresStorage) Update(sub *models.Subscription) error {
	log.Printf("Storage.Update: updating subscription with ID %d", sub.ID)
	log.Printf("Storage.Update: new data - service: '%s', price: %d, user_id: %s, start_date: %s, end_date: %s",
		sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate)

	result := s.db.Save(sub)
	if result.Error != nil {
		log.Printf("Storage.Update: failed to execute update for ID %d - %v", sub.ID, result.Error)
		return fmt.Errorf("failed to update: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("Storage.Update: no subscription found with ID %d", sub.ID)
		return fmt.Errorf("subscription with id %d not found", sub.ID)
	}

	log.Printf("Storage.Update: successfully updated subscription with ID %d (%d rows affected)",
		sub.ID, result.RowsAffected)
	return nil
}

func (s *PostgresStorage) GetAllEntries() ([]models.Subscription, error) {
	log.Printf("Storage.GetAllEntries: fetching all subscriptions from database")

	var subs []models.Subscription
	result := s.db.Find(&subs)
	if result.Error != nil {
		log.Printf("Storage.GetAllEntries: query execution failed - %v", result.Error)
		return nil, fmt.Errorf("query failed: %w", result.Error)
	}

	log.Printf("Storage.GetAllEntries: successfully retrieved %d subscription entries", len(subs))
	return subs, nil
}

func (s *PostgresStorage) GetOneEntry(id int) (*models.Subscription, error) {
	log.Printf("Storage.GetOneEntry: fetching subscription with ID %d", id)

	var sub models.Subscription
	result := s.db.Find(&sub, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("Storage.GetOneEntry: subscription not found with ID %d", id)
			return nil, fmt.Errorf("subscription with ID %d not found", id)
		}
		log.Printf("Storage.GetOneEntry: database error for ID %d - %v", id, result.Error)
		return nil, result.Error
	}

	log.Printf("Storage.GetOneEntry: found subscription with ID %d (service: '%s', price: %d, user_id: %s)",
		id, sub.ServiceName, sub.Price, sub.UserID)
	return &sub, nil
}

func (s *PostgresStorage) Delete(id int) error {
	log.Printf("Storage.Delete: attempting to delete subscription with ID %d", id)

	result := s.db.Delete(&models.Subscription{}, id)
	if result.Error != nil {
		log.Printf("Storage.Delete: failed to delete subscription with ID %d - %v", id, result.Error)
		return fmt.Errorf("failed to delete: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("Storage.Delete: no subscription found with ID %d", id)
		return fmt.Errorf("subscription with id %d not found", id)
	}

	log.Printf("Storage.Delete: successfully deleted subscription with ID %d (%d rows affected)", id, result.RowsAffected)
	return nil
}

func (s *PostgresStorage) Close() error {
	log.Printf("Storage.Close: closing database connection")

	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("Storage: get sql.DB failed: %w", err)
	}

	return sqlDB.Close()
}

func (s *PostgresStorage) SumPrice(start_date, end_date, ServiceName string, UserID uuid.UUID) (int, error) {
	log.Printf("Storage.SumPrice: calculating total price with filters")
	log.Printf("Storage.SumPrice: filters - user_id: %s, service_name: '%s', start_date: '%s', end_date: '%s'",
		UserID, ServiceName, start_date, end_date)

	query := s.db.Model(&models.Subscription{})

	if UserID != uuid.Nil {
		query = query.Where("user_id = ?", UserID)
	}

	if ServiceName != "" {
		query = query.Where("service_name = ?", ServiceName)
	}

	if start_date != "" {
		query = query.Where("start_date >= ?", start_date)
	}

	if end_date != "" {
		query = query.Where("end_date <= ?", end_date)
	}

	var total int
	result := s.db.Select("COALESCE(SUM(price), 0)").Scan(&total)
	if result.Error != nil {
		log.Printf("Storage.SumPrice: failed to calculate total price - %v", result.Error)
		return 0, fmt.Errorf("failed to calculate total price: %w", result.Error)
	}

	log.Printf("Storage.SumPrice: successfully calculated total price: %d rubles", total)
	return total, nil
}
