package rest

import (
	"net/http"

	"agent-michi/internal/domain"
	"agent-michi/internal/usecase"

	"github.com/gin-gonic/gin"
)

// Handler holds the use cases for HTTP handlers.
type Handler struct {
	deployer     domain.Deployer
	statsUseCase *usecase.StatsUseCase
}

// NewHandler creates a new Handler.
func NewHandler(deployer domain.Deployer, statsUseCase *usecase.StatsUseCase) *Handler {
	return &Handler{
		deployer:     deployer,
		statsUseCase: statsUseCase,
	}
}

// Deploy handles POST /api/v1/deploy.
func (h *Handler) Deploy(c *gin.Context) {
	var req domain.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.deployer.Deploy(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Stats handles GET /api/v1/stats.
func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.statsUseCase.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
