package controllers

import (
	"avitoTestTask/internal/models"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service userService
	router  *gin.Engine
	log     *slog.Logger
}

type userService interface {
	GetUserReviewPRs(userId string) ([]*models.PullRequest, error)
	SetUserActive(userId string, isActive bool) (*models.User, error)
}

func CreateUserController(service userService, router *gin.Engine, log *slog.Logger) UserController {
	return UserController{service: service, router: router, log: log}
}

func (h *UserController) EnableController() {
	h.router.POST("/users/setIsActive", h.UserSetIsActive)
	h.router.GET("/users/getReview", h.GetUserReviews)
}

func (h *UserController) UserSetIsActive(c *gin.Context) {
	const op = "handler.UserSetIsActive"

	var request struct {
		UserID   string `json:"user_id" binding:"required"`
		IsActive bool   `json:"is_active" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Error(op, " : ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	user, err := h.service.SetUserActive(request.UserID, request.IsActive)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		h.log.Error(op, " : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *UserController) GetUserReviews(c *gin.Context) {
	const op = "handler.GetUserReviews"

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "user_id parameter is required",
			},
		})
		return
	}

	pullRequests, err := h.service.GetUserReviewPRs(userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "User not found",
				},
			})
			return
		}
		h.log.Error(op, " : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": pullRequests,
	})
}
