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

func (s *PostgresStorage) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM subscriptions WHERE user_id = $1
	`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
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
