package handlers

import (
	"log"
	"net/http"
	"strconv"
	"subscription-aggregator/backend/internal/models"
	"subscription-aggregator/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubsHandler struct {
	storage *storage.PostgresStorage
}

func NewSubsHandler(storage *storage.PostgresStorage) *SubsHandler {
	log.Printf("Handlers: creating new SubsHandler")
	return &SubsHandler{
		storage: storage,
	}
}

func (h *SubsHandler) CreateEntry(c *gin.Context) {
	log.Printf("Handlers.CreateEntry: starting to create new subscription")

	var subscribe models.Subscription
	if err := c.BindJSON(&subscribe); err != nil {
		log.Printf("Handlers.CreateEntry: failed to bind JSON - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON input"})
		return
	}

	// Валидация
	if subscribe.ServiceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service name is required"})
		return
	}
	if subscribe.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be positive"})
		return
	}

	log.Printf("Handlers.CreateEntry: creating subscription for service '%s' with price %d, user_id %s, start date %s, end date %s",
		subscribe.ServiceName, subscribe.Price, subscribe.UserID, subscribe.StartDate, subscribe.EndDate)

	if err := h.storage.Create(&subscribe); err != nil {
		log.Printf("Handlers.CreateEntry: storage creation failed - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create entry"})
		return
	}

	log.Printf("Handlers.CreateEntry: successfully created subscription with ID %d for service '%s'",
		subscribe.ID, subscribe.ServiceName)
	c.JSON(http.StatusOK, subscribe)
}

func (h *SubsHandler) UpdateEntry(c *gin.Context) {
	log.Printf("Handlers.UpdateEntry: starting to update subscription")

	idStr := c.Param("id")
	log.Printf("Handlers.UpdateEntry: updating subscription with ID %s", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Handlers.UpdateEntry: invalid ID format '%s' - %v", idStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var subscribe models.Subscription
	if err := c.BindJSON(&subscribe); err != nil {
		log.Printf("Handlers.UpdateEntry: failed to bind JSON - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON input"})
		return
	}

	subscribe.ID = id

	log.Printf("Handlers.UpdateEntry: updating subscription with ID %d, new service '%s', price %d, user_id %s, start date %s, end date %s",
		subscribe.ID, subscribe.ServiceName, subscribe.Price, subscribe.UserID, subscribe.StartDate, subscribe.EndDate)

	if err := h.storage.Update(&subscribe); err != nil {
		log.Printf("Handlers.UpdateEntry: storage update failed for ID %d - %v", subscribe.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}

	log.Printf("Handlers.UpdateEntry: successfully updated subscription with ID %d", subscribe.ID)
	c.JSON(http.StatusOK, subscribe)
}

func (h *SubsHandler) GetAllEntries(c *gin.Context) {
	log.Printf("Handlers.GetAllEntries: fetching all subscriptions")

	entries, err := h.storage.GetAllEntries()
	if err != nil {
		log.Printf("Handlers.GetAllEntries: storage error - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Entries not found"})
		return
	}

	log.Printf("Handlers.GetAllEntries: successfully retrieved %d subscription entries", len(entries))
	c.JSON(http.StatusOK, entries)
}

func (h *SubsHandler) GetOneEntry(c *gin.Context) {
	idStr := c.Param("id")

	log.Printf("Handlers.GetOneEntry: fetching subscription with ID %s", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Handlers.GetOneEntry: Invalid ID parameter - %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	log.Printf("Handlers.GetOneEntry: looking for subscription with ID %d", id)

	entry, err := h.storage.GetOneEntry(id)
	if err != nil {
		log.Printf("Handlers.GetOneEntry: subscription not found with ID %d - %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	log.Printf("Handlers.GetOneEntry: successfully retrieved subscription with ID %d (service: '%s', user_id: %s)",
		id, entry.ServiceName, entry.UserID)
	c.JSON(http.StatusOK, entry)
}

func (h *SubsHandler) DeleteEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		log.Printf("Handlers.DeleteEntry: Invalid ID parameter - %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	log.Printf("Handlers.DeleteEntry: attempting to delete subscription with ID: %d", id)

	if err := h.storage.Delete(id); err != nil {
		log.Printf("Handlers.DeleteEntry: storage delete failed for ID %d - %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	log.Printf("Handlers.DeleteEntry: successfully deleted subscription with ID: %d", id)
	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted"})
}

func (h *SubsHandler) GetSumPrice(c *gin.Context) {
	log.Printf("Handlers.GetSumPrice: starting to calculate total price with filters")

	type Request struct {
		UserID      string `json:"user_id"`
		ServiceName string `json:"service_name"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		log.Printf("Handlers.GetSumPrice: failed to bind JSON request - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	log.Printf("Handlers.GetSumPrice: applying filters - user_id: '%s', service_name: '%s', start_date: '%s', end_date: '%s'",
		req.UserID, req.ServiceName, req.StartDate, req.EndDate)

	var userID uuid.UUID
	var err error

	if req.UserID != "" {
		userID, err = uuid.Parse(req.UserID)
		if err != nil {
			log.Printf("Handlers.GetSumPrice: invalid UUID format '%s' - %v", req.UserID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
			return
		}
		log.Printf("Handlers.GetSumPrice: parsed user_id filter: %s", userID)
	}

	sum, err := h.storage.SumPrice(req.StartDate, req.EndDate, req.ServiceName, userID)
	if err != nil {
		log.Printf("Handlers.GetSumPrice: storage sum calculation failed - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sum prices"})
		return
	}

	log.Printf("Handlers.GetSumPrice: successfully calculated total price: %d rubles", sum)
	c.JSON(http.StatusOK, gin.H{
		"total_price": sum,
		"filters": gin.H{
			"user_id":      req.UserID,
			"service_name": req.ServiceName,
			"start_date":   req.StartDate,
			"end_date":     req.EndDate,
		},
	})
}
