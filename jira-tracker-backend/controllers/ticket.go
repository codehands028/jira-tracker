package controllers

import (
	"net/http"
	"strconv"

	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// CreateTicket 创建工单
func CreateTicket(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req services.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置创建人ID
	req.CreatorID = userID.(uint)

	ticketService := services.NewTicketService()
	ticket, err := ticketService.CreateTicket(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// GetTicketList 获取工单列表
func GetTicketList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	priority := c.Query("priority")

	var currentUserID uint
	if userIDStr := c.Query("current_user_id"); userIDStr != "" {
		id, _ := strconv.ParseUint(userIDStr, 10, 32)
		currentUserID = uint(id)
	}

	ticketService := services.NewTicketService()
	tickets, total, err := ticketService.GetTicketList(page, pageSize, status, priority, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":      tickets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTicketDetail 获取工单详情
func GetTicketDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单ID"})
		return
	}

	ticketService := services.NewTicketService()
	ticket, flows, err := ticketService.GetTicketDetail(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ticket": ticket,
		"flows":  flows,
	})
}

// FlowTicket 流转工单
func FlowTicket(c *gin.Context) {
	userID, _ := c.Get("user_id")

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单ID"})
		return
	}

	var req struct {
		ToUserID uint   `json:"to_user_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flowReq := &services.FlowTicketRequest{
		TicketID:   uint(ticketID),
		FromUserID: userID.(uint),
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		Status:     "processing",
	}

	ticketService := services.NewTicketService()
	if err := ticketService.FlowTicket(flowReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "流转成功"})
}

// RetestTicket 复测工单
func RetestTicket(c *gin.Context) {
	userID, _ := c.Get("user_id")

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单ID"})
		return
	}

	var req struct {
		Passed   bool   `json:"passed" binding:"required"`
		Content  string `json:"content"`
		ToUserID uint   `json:"to_user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketService := services.NewTicketService()
	if req.Passed {
		// 复测通过，关闭工单
		if err := ticketService.CloseTicket(uint(ticketID), userID.(uint), req.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// 复测不通过，流转回研发
		if req.ToUserID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请指定处理人"})
			return
		}

		flowReq := &services.FlowTicketRequest{
			TicketID:   uint(ticketID),
			FromUserID: userID.(uint),
			ToUserID:   req.ToUserID,
			Content:    req.Content,
			Status:     "processing",
		}
		if err := ticketService.FlowTicket(flowReq); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// CloseTicket 关闭工单
func CloseTicket(c *gin.Context) {
	userID, _ := c.Get("user_id")

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单ID"})
		return
	}

	var req struct {
		Conclusion string `json:"conclusion" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketService := services.NewTicketService()
	if err := ticketService.CloseTicket(uint(ticketID), userID.(uint), req.Conclusion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "关闭成功"})
}

// BatchAssignTickets 批量分配工单
func BatchAssignTickets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req services.BatchAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OperatorID = userID.(uint)

	ticketService := services.NewTicketService()
	count, err := ticketService.BatchAssignTickets(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "批量分配完成",
		"total":         len(req.TicketIDs),
		"success_count": count,
	})
}

// BatchCloseTickets 批量关闭工单
func BatchCloseTickets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req services.BatchCloseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OperatorID = userID.(uint)

	ticketService := services.NewTicketService()
	count, err := ticketService.BatchCloseTickets(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "批量关闭完成",
		"total":         len(req.TicketIDs),
		"success_count": count,
	})
}
