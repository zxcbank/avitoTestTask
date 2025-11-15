package controllers

type teamController struct {
}

type teamService interface {
	CreateTeam(team *Team) error
	GetTeam(teamName string) (*Team, error)
}

func (h *teamController) TeamAdd(c *gin.Context) {

}

func (h *teamController) TeamGet(c *gin.Context) {

}
