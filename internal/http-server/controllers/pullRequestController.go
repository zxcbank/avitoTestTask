package controllers

import (
	"avitoTestTask/internal/models"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PullRequestController struct {
	service PullRequestService
	router  *gin.Engine
	log     *slog.Logger
}

type PullRequestService interface {
	CreatePullRequest(PullRequestId, PullRequestName, AuthorID string) (*models.PullRequest, error)
	GetPullRequest(PullRequestID string) (*models.PullRequest, error)
	MergePullRequest(PullRequestID string) (*models.PullRequest, error)
	ReassignReviewer(PullRequestID, OldUserId string) (*models.Reassign, error)
}

func CreatePullRequestController(service PullRequestService, router *gin.Engine, log *slog.Logger) PullRequestController {
	return PullRequestController{service: service, router: router, log: log}
}

func (h *PullRequestController) EnableController() {
	h.router.POST("/pullRequest/create", h.CreatePullRequest)
	h.router.POST("/pullRequest/merge", h.MergePullRequest)
	h.router.POST("/pullRequest/reassign", h.ReassignPullRequest)
}

func (h *PullRequestController) CreatePullRequest(c *gin.Context) {
	const op = "handler.CreatePullRequest"

	var request struct {
		PullRequestID   string `json:"pull_request_id" binding:"required"`
		PullRequestName string `json:"pull_request_name" binding:"required"`
		AuthorID        string `json:"author_id" binding:"required"`
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

	pr, err := h.service.CreatePullRequest(request.PullRequestID, request.PullRequestName, request.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound) || errors.Is(err, models.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "Author/team not found",
				},
			})
		case errors.Is(err, models.ErrPRExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": map[string]interface{}{
					"code":    "PR_EXISTS",
					"message": "PR id already exists",
				},
			})
		default:
			h.log.Error(op, " : ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]interface{}{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				},
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pr": pr,
	})
}

func (h *PullRequestController) MergePullRequest(c *gin.Context) {
	const op = "handler.MergePullRequest"

	var request struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
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

	pr, err := h.service.MergePullRequest(request.PullRequestID)
	if err != nil {
		if errors.Is(err, models.ErrPRNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "PR not found",
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
		"pr": pr,
	})
}

func (h *PullRequestController) ReassignPullRequest(c *gin.Context) {
	const op = "handler.ReassignPullRequest"

	var request struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
		OldUserID     string `json:"old_user_id" binding:"required"`
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

	pr, err := h.service.ReassignReviewer(request.PullRequestID, request.OldUserID)
	replacedBy := request.OldUserID
	if err != nil {
		switch {
		case errors.Is(err, models.ErrPRNotFound) || errors.Is(err, models.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "PR or user not found",
				},
			})
		case errors.Is(err, models.ErrPRMerged):
			c.JSON(http.StatusConflict, gin.H{
				"error": map[string]interface{}{
					"code":    "PR_MERGED",
					"message": "cannot reassign on merged PR",
				},
			})
		case errors.Is(err, models.ErrNotAssigned):
			c.JSON(http.StatusConflict, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_ASSIGNED",
					"message": "reviewer is not assigned to this PR",
				},
			})
		case errors.Is(err, models.ErrNoCandidate):
			c.JSON(http.StatusConflict, gin.H{
				"error": map[string]interface{}{
					"code":    "NO_CANDIDATE",
					"message": "no active replacement candidate in team",
				},
			})
		default:
			h.log.Error(op, " : ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]interface{}{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": replacedBy,
	})
}
