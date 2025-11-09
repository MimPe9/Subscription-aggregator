package handlers

import (
	"log"
	"net/http"
	"subscription-aggregator/backend/internal/models"
	"subscription-aggregator/backend/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubsHandler struct {
	storage *storage.PostgresStorage
}

func NewSubsHandler(storage *storage.PostgresStorage) *SubsHandler {
	return &SubsHandler{
		storage: storage,
	}
}

func (h *SubsHandler) CreateEntry(c *gin.Context) {
	log.Printf("Beginning of creation new subscription")

	var subscribe models.Subscription
	if err := c.BindJSON(&subscribe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON input"})
		return
	}

	if err := h.storage.Create(&subscribe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create entry"})
		return
	}

	log.Printf("Subscription created")
	c.JSON(http.StatusOK, subscribe)
}

func (h *SubsHandler) UpdateEntry(c *gin.Context) {
	log.Printf("Starting to update subscription")

	var subscribe models.Subscription
	if err := c.BindJSON(&subscribe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON input"})
		return
	}

	log.Printf("Updating subscription with ID %s", subscribe.UserId)

	if subscribe.UserId == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.storage.Update(&subscribe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}

	log.Printf("Successfully updated subscription with ID %s", subscribe.UserId)
	c.JSON(http.StatusOK, subscribe)
}

func (h *SubsHandler) GetAllEntries(c *gin.Context) {
	log.Printf("Fetching all subscriptions")

	entries, err := h.storage.GetAllEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Entries not found"})
		return
	}

	log.Printf("Successfully retrieved %d entries", len(entries))
	c.JSON(http.StatusOK, entries)
}

func (h *SubsHandler) GetOneEntry(c *gin.Context) {
	idStr := c.Param("user_id")

	log.Printf("Fetching subscription with ID %s", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	entry, err := h.storage.GetOneEntry(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	log.Printf("Successfully retrieved subscription with ID %s", id)
	c.JSON(http.StatusOK, entry)
}

func (h *SubsHandler) DeleteEntry(c *gin.Context) {
	idStr := c.Param("user_id")

	log.Printf("Delete request for user_id: %s", idStr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	log.Printf("Attempting to delete UUID: %s", id)

	if err := h.storage.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}

	log.Printf("Successfully deleted: %s", id)
	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted"})
}

func (h *SubsHandler) GetSumPrice(c *gin.Context) {
	log.Printf("Calculating total price with filters")

	type Request struct {
		UserID      string `json:"user_id"`
		ServiceName string `json:"service_name"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	log.Printf("Filters - user_id: '%s', service_name: '%s', start_date: '%s', end_date: '%s'",
		req.UserID, req.ServiceName, req.StartDate, req.EndDate)

	var id uuid.UUID
	var err error

	if req.UserID != "" {
		id, err = uuid.Parse(req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
			return
		}
	}

	sum, err := h.storage.SumPrice(req.StartDate, req.EndDate, req.ServiceName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sum prices"})
		return
	}

	log.Printf("Calculated total price: %d", sum)
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
