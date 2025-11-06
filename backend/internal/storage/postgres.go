package storage

import (
	"database/sql"
	"fmt"
	"log"
	"subscription-aggregator/backend/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	log.Printf("Connection string: %s", connStr)

	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Attempt %d: sql.Open error: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err == nil {
			log.Println("Successfully connected to database")
			return &PostgresStorage{db: db}, nil
		}
		log.Printf("Attempt %d: Failed to ping database: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
}

func (s *PostgresStorage) Create(sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, start_date)
		VALUES ($1, $2, $3)
		RETURNING user_id
	`

	return s.db.QueryRow(query, sub.ServiceName, sub.Price, sub.StartDate).Scan(&sub.UserId)
}

func (s *PostgresStorage) Update(sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3
		WHERE user_id = $4
	`

	res, err := s.db.Exec(query, sub.ServiceName, sub.Price, sub.StartDate, sub.UserId)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription with id %s not found", sub.UserId)
	}

	return nil
}

func (s *PostgresStorage) GetAllEntries() ([]models.Subscription, error) {
	query := `
		SELECT service_name, price, user_id, start_date 
		FROM subscriptions
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(&sub.ServiceName, &sub.Price, &sub.UserId, &sub.StartDate)
		if err != nil {
			return nil, err
		}
		entries = append(entries, sub)
	}

	return entries, nil
}

func (s *PostgresStorage) GetOneEntry(id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT service_name, price, user_id, start_date
		FROM subscriptions
		WHERE user_id = $1
	`

	var entry models.Subscription
	err := s.db.QueryRow(query, id).Scan(&entry.ServiceName, &entry.Price, &entry.UserId, &entry.StartDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription with user_id %s not found", id)
		}
		return nil, err
	}
	return &entry, nil
}

func (s *PostgresStorage) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM subscriptions WHERE user_id = $1
	`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription with id %d not found", id)
	}
	return nil
}

func (s *PostgresStorage) Close() error {
	if s.db != nil {
		return s.Close()
	}
	return nil
}

/*func (s *PostgresStorage) SumPrice(start, end, ServiceName string, UserId uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE start_date >= $1 AND start_date <= $2
	`
}*/
