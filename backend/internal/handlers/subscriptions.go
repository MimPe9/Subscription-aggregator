package handlers

import (
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
	var subscribe models.Subscription
	if err := c.BindJSON(&subscribe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.Create(&subscribe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscribe)
}

func (h *SubsHandler) UpdateEntry(c *gin.Context) {
	var subscribe models.Subscription
	if err := c.BindJSON(&subscribe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if subscribe.UserId == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.storage.Update(&subscribe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscribe)
}

func (h *SubsHandler) GetAllEntries(c *gin.Context) {
	entries, err := h.storage.GetAllEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *SubsHandler) GetOneEntry(c *gin.Context) {
	idStr := c.Param("user_id")

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

	c.JSON(http.StatusOK, entry)
}

func (h *SubsHandler) DeleteEntry(c *gin.Context) {
	idStr := c.Param("user_id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
	}
	if err := h.storage.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Entry deleted"})
}
