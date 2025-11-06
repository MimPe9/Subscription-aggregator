package main

import (
	"fmt"
	"log"
	"subscription-aggregator/backend/config"
	"subscription-aggregator/backend/internal/handlers"
	"subscription-aggregator/backend/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	s := setupStorage()
	defer s.Close()

	r := gin.Default()
	r.Use(cors.Default())

	r.Static("/static", "./frontend")

	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/frontend.html")
	})

	subsHandler := handlers.NewSubsHandler(s)

	api := r.Group("/api")
	{
		api.GET("/subscriptions", subsHandler.GetAllEntries)
		api.POST("/subscriptions", subsHandler.CreateEntry)
		api.DELETE("/subscriptions/del/:user_id", subsHandler.DeleteEntry)
		api.PUT("/subscriptions/:user_id", subsHandler.UpdateEntry)
		api.GET("/subscriptions/:user_id", subsHandler.GetOneEntry)
	}

	r.Run(":8010")
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
