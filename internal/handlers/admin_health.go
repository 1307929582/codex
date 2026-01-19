package handlers

import (
	"net/http"

	"codex-gateway/internal/upstream"

	"github.com/gin-gonic/gin"
)

// AdminTriggerHealthCheck manually triggers health check for all upstreams
func AdminTriggerHealthCheck(c *gin.Context) {
	// This will be done asynchronously
	go upstream.GetHealthChecker().CheckAllUpstreams()

	c.JSON(http.StatusOK, gin.H{
		"message": "Health check triggered for all upstreams",
	})
}

// AdminGetUpstreamHealth gets health status of all upstreams
func AdminGetUpstreamHealth(c *gin.Context) {
	upstreams := upstream.GetSelector().GetAllUpstreams()
	checker := upstream.GetHealthChecker()

	type UpstreamHealth struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Status        string `json:"status"`
		FailureCount  int    `json:"failure_count"`
		LastChecked   string `json:"last_checked"`
	}

	healthStatus := make([]UpstreamHealth, 0, len(upstreams))
	for _, u := range upstreams {
		lastChecked := "never"
		if u.LastChecked != nil {
			lastChecked = u.LastChecked.Format("2006-01-02 15:04:05")
		}

		healthStatus = append(healthStatus, UpstreamHealth{
			ID:           u.ID,
			Name:         u.Name,
			Status:       u.Status,
			FailureCount: checker.GetFailureCount(u.ID),
			LastChecked:  lastChecked,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"upstreams": healthStatus,
	})
}
