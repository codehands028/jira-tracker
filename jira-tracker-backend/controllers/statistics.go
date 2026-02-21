package controllers

import (
	"net/http"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// GetStatistics 获取统计数据
func GetStatistics(c *gin.Context) {
	ticketService := services.NewTicketService()
	stats, err := ticketService.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboard 获取数据看板
func GetDashboard(c *gin.Context) {
	ticketService := services.NewTicketService()
	data, err := ticketService.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
