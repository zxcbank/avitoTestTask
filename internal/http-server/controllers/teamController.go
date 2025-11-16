package controllers

import (
	"avitoTestTask/internal/models"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TeamController struct {
	service teamService
	router  *gin.Engine
	log     *slog.Logger
}

type teamService interface {
	CreateTeam(team *models.Team) error
	GetTeam(teamName string) (*models.Team, error)
}

func CreateTeamController(service teamService, router *gin.Engine, log *slog.Logger) TeamController {
	return TeamController{service: service, router: router, log: log}
}

func (h *TeamController) EnableController() {
	h.router.GET("/team/get", h.TeamGet)
	h.router.POST("/team/add", h.CreateTeam)
}

func (h *TeamController) CreateTeam(c *gin.Context) {
	const op = "handler.CreateTeam"

	var request models.Team
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

	err := h.service.CreateTeam(&request)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrTeamExists):
			h.log.Error(op, " : ", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": map[string]interface{}{
					"code":    "TEAM_EXISTS",
					"message": "team_name already exists",
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

	h.log.Info(op, " : ", "team created", request.Name)
	c.JSON(http.StatusCreated, gin.H{
		"team": request,
	})
}

func (h *TeamController) TeamGet(c *gin.Context) {
	const op = "handler.TeamGet"

	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"code":    "INVALID_REQUEST",
				"message": "team_name is required",
			},
		})
		return
	}

	team, err := h.service.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, models.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": map[string]interface{}{
					"code":    "NOT_FOUND",
					"message": "Team not found",
				},
			})
			return
		}
		h.log.Error(op, " : ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
		})
		return
	}

	h.log.Info(op, " : ", "team found", team.Name)
	c.JSON(http.StatusOK, team)
}
