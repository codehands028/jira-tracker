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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误，请检查输入内容"})
		return
	}

	// 设置创建人ID
	req.CreatorID = userID.(uint)

	ticketService := services.NewTicketService()
	ticket, err := ticketService.CreateTicket(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建工单失败，请稍后重试"})
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

	// 从JWT token中获取用户信息
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var currentUserID uint
	
	// 根据角色决定可见范围
	// 管理员可以看到所有工单，普通用户只能看到自己的工单
	if role.(string) != "admin" {
		// 非管理员：强制只显示自己的工单
		currentUserID = userID.(uint)
	} else {
		// 管理员：可以根据查询参数筛选
		if userIDStr := c.Query("current_user_id"); userIDStr != "" {
			id, _ := strconv.ParseUint(userIDStr, 10, 32)
			currentUserID = uint(id)
		}
	}

	// 解析超时筛选参数
	var isTimeout *bool
	if isTimeoutStr := c.Query("is_timeout"); isTimeoutStr != "" {
		val := isTimeoutStr == "true"
		isTimeout = &val
	}

	ticketService := services.NewTicketService()
	tickets, total, err := ticketService.GetTicketList(page, pageSize, status, priority, currentUserID, isTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取工单列表失败，请稍后重试"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取工单详情失败，请稍后重试"})
		return
	}

	// 获取SLA状态
	slaStatus, _ := ticketService.GetTicketSLAStatus(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"ticket":     ticket,
		"flows":      flows,
		"sla_status": slaStatus,
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

	// 记录批量操作日志
	detail := &services.BatchOperationDetail{
		ToUserID: req.ToUserID,
		Content:  req.Content,
	}
	services.CreateBatchOperationLog(userID.(uint), "assign", req.TicketIDs, count, detail, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":       "批量分配完成",
		"total":         len(req.TicketIDs),
		"success_count": count,
	})
}

// BatchFlowTickets 批量流转工单
func BatchFlowTickets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	role, _ := c.Get("role")

	var req services.BatchFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OperatorID = userID.(uint)
	req.UserRole = role.(string)

	ticketService := services.NewTicketService()
	count, err := ticketService.BatchFlowTickets(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录批量操作日志
	detail := &services.BatchOperationDetail{
		ToUserID: req.ToUserID,
		Content:  req.Content,
	}
	services.CreateBatchOperationLog(userID.(uint), "flow", req.TicketIDs, count, detail, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":       "批量流转完成",
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

	role, _ := c.Get("role")

	var req services.BatchCloseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.OperatorID = userID.(uint)
	req.UserRole = role.(string)

	ticketService := services.NewTicketService()
	count, err := ticketService.BatchCloseTickets(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 记录批量操作日志
	detail := &services.BatchOperationDetail{
		Conclusion: req.Conclusion,
	}
	services.CreateBatchOperationLog(userID.(uint), "close", req.TicketIDs, count, detail, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":       "批量关闭完成",
		"total":         len(req.TicketIDs),
		"success_count": count,
	})
}
