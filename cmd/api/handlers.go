package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zepheo/guild_manager/internal/service"
)

// RaidHandler acts as a container for our dependencies
type RaidHandler struct {
	Service *service.RaidService
}

// NewRaidHandler creates a new instance with the injected service
func NewRaidHandler(s *service.RaidService) *RaidHandler {
	return &RaidHandler{Service: s}
}

func (h *RaidHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/raids/upload", h.HandleRaidUpload)
	r.GET("/health", h.HandleHealth)
}

func (h *RaidHandler) HandleRaidUpload(c *gin.Context) {
	logID := c.PostForm("log_id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "log_id is required"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "csv file is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	// Use the injected service instead of a global variable
	err = h.Service.ProcessRaid(c.Request.Context(), f, logID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Raid processed successfully"})
}

func (h *RaidHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
