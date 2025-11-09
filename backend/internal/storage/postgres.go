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
	log.Printf("Successfully created subscription with ID %s", sub.UserId)
	return s.db.QueryRow(query, sub.ServiceName, sub.Price, sub.StartDate).Scan(&sub.UserId)
}

func (s *PostgresStorage) Update(sub *models.Subscription) error {
	log.Printf("Updating subscription with ID %s", sub.UserId)

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

	log.Printf("Successfully updated subscription with ID %s", sub.UserId)
	return nil
}

func (s *PostgresStorage) GetAllEntries() ([]models.Subscription, error) {
	log.Printf("Fetching all subscriptions")

	query := `
		SELECT service_name, price, user_id, start_date 
		FROM subscriptions
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var entries []models.Subscription
	count := 0
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(&sub.ServiceName, &sub.Price, &sub.UserId, &sub.StartDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan data: %w", err)
		}
		entries = append(entries, sub)
		count++
	}

	log.Printf("Successfully retrieved %d entries", count)
	return entries, nil
}

func (s *PostgresStorage) GetOneEntry(id uuid.UUID) (*models.Subscription, error) {
	log.Printf("Fetching subscription with ID %s", id)

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

	log.Printf("Successfully retrieved subscription with ID %s", id)
	return &entry, nil
}

func (s *PostgresStorage) Delete(id uuid.UUID) error {
	log.Printf("Deleting subscription with ID %s", id)

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
		return fmt.Errorf("subscription with id %s not found", id)
	}

	log.Printf("Successfully deleted subscription with ID %s", id)
	return nil
}

func (s *PostgresStorage) Close() error {
	log.Printf("Closing database connection")

	if s.db != nil {
		return s.db.Close()
	}

	log.Printf("Database connection closed successfully")
	return nil
}

func (s *PostgresStorage) SumPrice(start, end, ServiceName string, UserId uuid.UUID) (int, error) {
	log.Printf("Calculating total price")

	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
          AND ($2::text IS NULL OR service_name = $2)
          AND ($3::text IS NULL OR start_date >= $3)
          AND ($4::text IS NULL OR start_date <= $4)
	`
	var UserIdParam interface{} = nil
	if UserId != uuid.Nil {
		UserIdParam = UserId
	}

	var ServiceNameParam interface{} = nil
	if ServiceName != "" {
		ServiceNameParam = ServiceName
	}

	var startParam interface{} = nil
	if start != "" {
		startParam = start
	}

	var endParam interface{} = nil
	if end != "" {
		endParam = end
	}

	log.Printf("Executing query with params - user_id: %v, service: %v, start: %v, end: %v",
		UserIdParam, ServiceNameParam, startParam, endParam)

	var total int
	err := s.db.QueryRow(query, UserIdParam, ServiceNameParam, startParam, endParam).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total price: %w", err)
	}

	log.Printf("Calculated total price: %d", total)
	return total, nil
}
