package controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"
	"jira-tracker-backend/services"

	"github.com/gin-gonic/gin"
)

// ExportExcel 导出Excel（异步方式）
func ExportExcel(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	var req services.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.UserID = userID
	req.UserRole = userRole

	exportService := services.NewExportService()

	// 创建导出任务
	task, err := exportService.CreateExportTask(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建导出任务失败"})
		return
	}

	// 异步执行导出
	go exportService.ExportToExcelAsync(&req, task.ID)

	// 记录操作日志
	database.DB.Create(&models.OperationLog{
		UserID:      userID,
		Module:      "export",
		Action:      "POST",
		Description: "创建Excel导出任务",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "导出任务已创建，请稍后下载",
		"task_id": task.ID,
	})
}

// DownloadExcel 下载Excel文件
func DownloadExcel(c *gin.Context) {
	userID := c.GetUint("user_id")
	filename := c.Query("filename")

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件名不能为空"})
		return
	}

	// 权限检查：验证该文件是否属于当前用户
	var task models.ExportTask
	err := database.DB.Where("filename = ? AND user_id = ?", filename, userID).First(&task).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权下载此文件"})
		return
	}

	// 使用配置的存储路径
	storagePath := "/tmp" // 默认路径
	if config.GlobalConfig != nil && config.GlobalConfig.Export.StoragePath != "" {
		storagePath = config.GlobalConfig.Export.StoragePath
	}
	filepath := storagePath + "/" + filename

	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 记录操作日志
	database.DB.Create(&models.OperationLog{
		UserID:      userID,
		Module:      "export",
		Action:      "GET",
		Description: "下载Excel文件: " + filename,
	})

	c.FileAttachment(filepath, filename)
}

// GetPersonalDashboard 获取个人数据看板
func GetPersonalDashboard(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")
	priority := c.Query("priority")
	timeDimension := c.Query("time_dimension")

	exportService := services.NewExportService()
	data, err := exportService.GetPersonalDashboard(userID, userRole, startDate, endDate, status, priority, timeDimension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取个人数据看板失败"})
		return
	}

	c.JSON(http.StatusOK, data)
}

// ExportTicketsDirect 直接导出工单（同步方式，适合数据量较小的情况）
func ExportTicketsDirect(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	var req services.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.UserID = userID
	req.UserRole = userRole

	exportService := services.NewExportService()

	// 创建导出任务记录（用于下载权限验证）
	task, err := exportService.CreateExportTask(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建导出任务失败"})
		return
	}

	// 同步导出不需要进度回调
	filepath, err := exportService.ExportToExcel(&req, nil)
	if err != nil {
		// 更新任务状态为失败
		task.Status = "failed"
		task.Error = err.Error()
		exportService.UpdateExportTask(task)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败: " + err.Error()})
		return
	}

	// 获取文件名（从完整路径中提取）
	storagePath := "/tmp" // 默认路径
	if config.GlobalConfig != nil && config.GlobalConfig.Export.StoragePath != "" {
		storagePath = config.GlobalConfig.Export.StoragePath
	}
	filename := filepath[len(storagePath)+1:]

	// 更新任务状态为完成，保存文件名用于下载验证
	now := time.Now()
	task.Status = "completed"
	task.FilePath = filepath
	task.Filename = filename
	task.CompletedAt = &now
	exportService.UpdateExportTask(task)

	// 记录操作日志
	database.DB.Create(&models.OperationLog{
		UserID:      userID,
		Module:      "export",
		Action:      "POST",
		Description: "导出工单Excel",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":      "导出成功",
		"filename":     filename,
		"download_url": "/api/export/download?filename=" + filename,
	})
}

// GetExportTaskStatus 获取导出任务状态
func GetExportTaskStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	exportService := services.NewExportService()
	task, err := exportService.GetExportTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 检查权限：只能查看自己的任务
	if task.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权查看此任务"})
		return
	}

	response := gin.H{
		"task_id":       task.ID,
		"status":        task.Status,
		"total_count":   task.TotalCount,
		"process_count": task.ProcessCount,
	}

	if task.Status == "completed" {
		response["filename"] = task.Filename
		response["download_url"] = "/api/export/download?filename=" + task.Filename
	} else if task.Status == "failed" {
		response["error"] = task.Error
	}

	c.JSON(http.StatusOK, response)
}

// GetUserConfig 获取用户配置
func GetUserConfig(c *gin.Context) {
	userID := c.GetUint("user_id")

	exportService := services.NewExportService()
	config, err := exportService.GetUserConfig(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dashboard_config": config.DashboardConfig,
	})
}

// UpdateUserConfig 更新用户配置
func UpdateUserConfig(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		DashboardConfig string `json:"dashboard_config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	exportService := services.NewExportService()
	if err := exportService.UpdateUserConfig(userID, req.DashboardConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}

// GetDashboardTickets 获取看板详情工单列表（点击图表后查看）
func GetDashboardTickets(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	date := c.Query("date")
	status := c.Query("status")
	priority := c.Query("priority")
	ticketType := c.Query("ticket_type")
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	exportService := services.NewExportService()
	tickets, total, err := exportService.GetDashboardTickets(userID, userRole, date, status, priority, ticketType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取工单列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  tickets,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ExportTicketsAsync 异步导出（带进度反馈）
func ExportTicketsAsync(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	var req services.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	req.UserID = userID
	req.UserRole = userRole

	exportService := services.NewExportService()

	// 创建导出任务
	task, err := exportService.CreateExportTask(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建导出任务失败"})
		return
	}

	// 异步执行导出
	go func() {
		// 更新任务状态为处理中
		task.Status = "processing"
		task.UpdatedAt = time.Now()
		exportService.UpdateExportTask(task)

		// 执行导出（使用进度回调更新任务进度）
		filepath, err := exportService.ExportToExcel(&req, func(processed, total int) {
			task.TotalCount = total
			task.ProcessCount = processed
			task.UpdatedAt = time.Now()
			exportService.UpdateExportTask(task)
		})
		
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			task.UpdatedAt = time.Now()
			exportService.UpdateExportTask(task)
			return
		}

		// 更新任务状态为完成
		task.Status = "completed"
		task.FilePath = filepath
		// 从完整路径提取文件名
		storagePath := "/tmp"
		if config.GlobalConfig != nil && config.GlobalConfig.Export.StoragePath != "" {
			storagePath = config.GlobalConfig.Export.StoragePath
		}
		task.Filename = filepath[len(storagePath)+1:]
		task.TotalCount = task.ProcessCount // 确保总数正确
		now := time.Now()
		task.CompletedAt = &now
		task.UpdatedAt = now
		exportService.UpdateExportTask(task)
	}()

	// 记录操作日志
	database.DB.Create(&models.OperationLog{
		UserID:      userID,
		Module:      "export",
		Action:      "POST",
		Description: "创建异步Excel导出任务",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "导出任务已创建",
		"task_id": task.ID,
	})
}
