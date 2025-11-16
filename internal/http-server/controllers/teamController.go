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
	h.router.GET("/team/get/:team_name", h.TeamGet)
	h.router.GET("/team/add", h.CreateTeam)
}

func (h *TeamController) CreateTeam(c *gin.Context) {
	const op = "handler.CreateTeam"

	var request models.Team
	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Error(op, " : ", err.Error())

		c.JSON(http.StatusBadRequest, gin.H{"INVALID_REQUEST": "Invalid request body"})
		return
	}

	err := h.service.CreateTeam(&request)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrTeamExists):
			h.log.Error(op, " : ", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"TEAM_EXISTS": "team_name already exists"})
		default:
			h.log.Error(op, " : ", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"INTERNAL_ERROR": "Internal server error"})
		}
		return
	}

	h.log.Info(op, " : ", "team created", request.Name)
	c.JSON(http.StatusCreated, gin.H{
		"team": request,
	})
}

func (h *TeamController) TeamGet(c *gin.Context) {
	if c.Request.Method != "GET" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "must use \"GET\""})
		return
	}

	teamName := c.Param("team_name")
	team, err := h.service.GetTeam(teamName)
	switch err {
	case nil:
		c.JSON(http.StatusOK, team)
	case models.ErrEmptyTeamName:
		c.JSON(http.StatusBadRequest, gin.H{"error": models.ErrTeamExists.Error()})
	case models.ErrTeamNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": models.ErrTeamNotFound.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

}
