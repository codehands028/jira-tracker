package controllers

import (
	"net/http"
	"strconv"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// GetTimeoutRules 获取超时规则列表
func GetTimeoutRules(c *gin.Context) {
	ticketService := services.NewTimeoutRuleService()
	rules, err := ticketService.GetTimeoutRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"list": rules})
}

// GetActiveTimeoutRule 获取当前生效的超时规则
func GetActiveTimeoutRule(c *gin.Context) {
	ticketService := services.NewTimeoutRuleService()
	rule, err := ticketService.GetActiveTimeoutRule()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateTimeoutRule 创建超时规则
func CreateTimeoutRule(c *gin.Context) {
	var req services.CreateTimeoutRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketService := services.NewTimeoutRuleService()
	rule, err := ticketService.CreateTimeoutRule(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateTimeoutRule 更新超时规则
func UpdateTimeoutRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req services.UpdateTimeoutRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketService := services.NewTimeoutRuleService()
	rule, err := ticketService.UpdateTimeoutRule(uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// SetActiveTimeoutRule 设置生效的超时规则
func SetActiveTimeoutRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ticketService := services.NewTimeoutRuleService()
	if err := ticketService.SetActiveTimeoutRule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置成功"})
}

// SetInactiveTimeoutRule 禁用超时规则
func SetInactiveTimeoutRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ticketService := services.NewTimeoutRuleService()
	if err := ticketService.SetInactiveTimeoutRule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "禁用成功"})
}

// DeleteTimeoutRule 删除超时规则
func DeleteTimeoutRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ticketService := services.NewTimeoutRuleService()
	if err := ticketService.DeleteTimeoutRule(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
