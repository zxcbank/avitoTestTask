package controllers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	router *gin.Engine
	log    *slog.Logger
}

func CreateHealthController(router *gin.Engine, log *slog.Logger) *HealthController {
	return &HealthController{router: router, log: log}
}

func (h *HealthController) EnableController() {
	h.router.GET("/health", h.HealthCheck)
}

func (h *HealthController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
