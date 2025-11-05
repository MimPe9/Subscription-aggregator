package main

import (
	"fmt"
	"log"
	"subscription-aggregator/backend/config"
	"subscription-aggregator/backend/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	s := setupStorage()
	defer s.Close()

	r := gin.Default()
	r.Use(cors.Default())
}

func setupStorage() *storage.PostgresStorage {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=%s",
		config.GetDBUser(), config.GetDBName(), config.GetDBPass(),
		config.GetDBHost(), config.GetDBPort(), config.GetDBSSLMode())

	log.Printf("Connecting to database: %s@%s:%s", config.GetDBUser(), config.GetDBHost(), config.GetDBPort())

	s, err := storage.NewPostgresStorage(connStr)
	if err != nil {
		log.Fatal("can't connect to storage: ", err)
	}

	return s
}
