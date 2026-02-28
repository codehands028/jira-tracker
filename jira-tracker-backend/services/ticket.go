package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"gorm.io/gorm"
)

type TicketService struct{}

func NewTicketService() *TicketService {
	return &TicketService{}
}

// CreateTicket 创建工单
func (s *TicketService) CreateTicket(req *CreateTicketRequest) (*models.Ticket, error) {
	// 验证处理人是否存在
	var user models.User
	if err := database.DB.First(&user, req.CurrentUserID).Error; err != nil {
		return nil, errors.New("指定的处理人不存在")
	}

	ticket := &models.Ticket{
		JiraKey:       req.JiraKey,
		JiraURL:       req.JiraURL,
		Description:   req.Description,
		Priority:      req.Priority,
		Status:        "processing",
		CurrentUserID: req.CurrentUserID,
		IsTimeout:     false,
	}

	if err := database.DB.Create(ticket).Error; err != nil {
		return nil, fmt.Errorf("创建工单失败: %w", err)
	}

	// 创建第一条流转记录
	flow := &models.TicketFlow{
		TicketID:   ticket.ID,
		FromUserID: req.CreatorID,
		ToUserID:   req.CurrentUserID,
		Content:    "创建工单",
		IsTimeout:  false,
	}

	if err := database.DB.Create(flow).Error; err != nil {
		return nil, fmt.Errorf("创建流转记录失败: %w", err)
	}

	// 更新用户待处理工单数
	database.DB.Model(&user).UpdateColumn("ticket_num", gorm.Expr("ticket_num + ?", 1))

	return ticket, nil
}

// FlowTicket 工单流转
func (s *TicketService) FlowTicket(req *FlowTicketRequest) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 获取工单信息
		var ticket models.Ticket
		if err := tx.First(&ticket, req.TicketID).Error; err != nil {
			return errors.New("工单不存在")
		}

		// 验证当前处理人
		if ticket.CurrentUserID != req.FromUserID {
			return errors.New("只有当前处理人可以流转工单")
		}

		// 获取上一条流转记录
		var lastFlow models.TicketFlow
		if err := tx.Where("ticket_id = ?", ticket.ID).Order("created_at desc").First(&lastFlow).Error; err != nil {
			return errors.New("获取流转记录失败")
		}

		// 计算处理时长
		processTime := time.Since(lastFlow.CreatedAt)

		// 检查是否超时
		var rule models.TimeoutRule
		isTimeout := false
		timeoutLevel := ""
		if err := tx.Where("is_active = ?", true).First(&rule).Error; err == nil {
			isTimeout = processTime > rule.NormalLimit
			if isTimeout {
				if processTime > rule.SevereLimit {
					timeoutLevel = "severe"
				} else {
					timeoutLevel = "normal"
				}
			}
		}

		// 创建新的流转记录
		newFlow := &models.TicketFlow{
			TicketID:     ticket.ID,
			FromUserID:   req.FromUserID,
			ToUserID:     req.ToUserID,
			Content:      req.Content,
			ProcessTime:  processTime,
			IsTimeout:    isTimeout,
			TimeoutLevel: timeoutLevel,
		}
		if err := tx.Create(newFlow).Error; err != nil {
			return fmt.Errorf("创建流转记录失败: %w", err)
		}

		// 更新工单状态
		updates := map[string]any{
			"current_user_id": req.ToUserID,
			"is_timeout":      false,
			"timeout_level":   "",
		}

		if req.Status != "" {
			updates["status"] = req.Status
		}

		if err := tx.Model(&ticket).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新工单失败: %w", err)
		}

		// 更新原处理人待处理工单数
		tx.Model(&models.User{}).Where("id = ?", req.FromUserID).
			UpdateColumn("ticket_num", gorm.Expr("ticket_num - ?", 1))

		// 更新新处理人待处理工单数
		tx.Model(&models.User{}).Where("id = ?", req.ToUserID).
			UpdateColumn("ticket_num", gorm.Expr("ticket_num + ?", 1))

		return nil
	})
}

// CloseTicket 关闭工单
func (s *TicketService) CloseTicket(ticketID uint, testerID uint, conclusion string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var ticket models.Ticket
		if err := tx.First(&ticket, ticketID).Error; err != nil {
			return errors.New("工单不存在")
		}

		if ticket.Status != "retesting" && ticket.Status != "processing" {
			return errors.New("只有处理中或待复测状态的工单才能关闭")
		}

		// 获取上一条流转记录
		var lastFlow models.TicketFlow
		if err := tx.Where("ticket_id = ?", ticket.ID).Order("created_at desc").First(&lastFlow).Error; err != nil {
			return errors.New("获取流转记录失败")
		}

		// 计算处理时长
		processTime := time.Since(lastFlow.CreatedAt)

		// 根据工单状态设置流转内容
		flowContent := ""
		if ticket.Status == "retesting" {
			flowContent = fmt.Sprintf("复测结论: %s", conclusion)
		} else {
			flowContent = fmt.Sprintf("关闭工单: %s", conclusion)
		}

		// 创建流转记录
		flow := &models.TicketFlow{
			TicketID:    ticket.ID,
			FromUserID:  ticket.CurrentUserID,
			ToUserID:    testerID,
			Content:     flowContent,
			ProcessTime: processTime,
			IsTimeout:   ticket.IsTimeout,
		}
		if err := tx.Create(flow).Error; err != nil {
			return fmt.Errorf("创建流转记录失败: %w", err)
		}

		// 更新工单状态
		if err := tx.Model(&ticket).Updates(map[string]any{
			"status": "closed",
		}).Error; err != nil {
			return fmt.Errorf("更新工单失败: %w", err)
		}

		// 更新当前处理人待处理工单数
		tx.Model(&models.User{}).Where("id = ?", ticket.CurrentUserID).
			UpdateColumn("ticket_num", gorm.Expr("ticket_num - ?", 1))

		return nil
	})
}

// GetTicketList 获取工单列表
func (s *TicketService) GetTicketList(page, pageSize int, status, priority string, currentUserID uint) ([]models.Ticket, int64, error) {
	var tickets []models.Ticket
	var total int64

	query := database.DB.Model(&models.Ticket{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if currentUserID > 0 {
		query = query.Where("current_user_id = ?", currentUserID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("CurrentUser").Offset(offset).Limit(pageSize).Order("created_at desc").Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// GetTicketDetail 获取工单详情
func (s *TicketService) GetTicketDetail(ticketID uint) (*models.Ticket, []models.TicketFlow, error) {
	var ticket models.Ticket
	if err := database.DB.Preload("CurrentUser").First(&ticket, ticketID).Error; err != nil {
		return nil, nil, errors.New("工单不存在")
	}

	var flows []models.TicketFlow
	if err := database.DB.Where("ticket_id = ?", ticketID).
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at asc").
		Find(&flows).Error; err != nil {
		return nil, nil, err
	}

	return &ticket, flows, nil
}

// CheckTimeout 检查工单是否超时
func (s *TicketService) CheckTimeout() error {
	var rule models.TimeoutRule
	if err := database.DB.Where("is_active = ?", true).First(&rule).Error; err != nil {
		return errors.New("未找到有效的超时规则")
	}

	// 获取所有处理中的工单
	var tickets []models.Ticket
	if err := database.DB.Where("status = ? AND is_timeout = ?", "processing", false).Find(&tickets).Error; err != nil {
		return fmt.Errorf("获取工单列表失败: %w", err)
	}

	for _, ticket := range tickets {
		// 获取工单最后流转记录
		var lastFlow models.TicketFlow
		if err := database.DB.Where("ticket_id = ?", ticket.ID).Order("created_at desc").First(&lastFlow).Error; err != nil {
			continue
		}

		// 计算停留时长
		stayTime := time.Since(lastFlow.CreatedAt)

		// 检查是否超时
		if stayTime > rule.NormalLimit {
			timeoutLevel := "normal"
			if stayTime > rule.SevereLimit {
				timeoutLevel = "severe"
			}

			// 更新工单超时状态
			database.DB.Model(&ticket).Updates(map[string]any{
				"is_timeout":    true,
				"timeout_level": timeoutLevel,
				"status":        "timeout",
			})

			// 创建超时通知
			notification := &models.Notification{
				UserID:   ticket.CurrentUserID,
				Type:     "timeout",
				Title:    "工单超时提醒",
				Content:  fmt.Sprintf("工单 %s 已超时，请尽快处理", ticket.JiraKey),
				TicketID: ticket.ID,
			}
			database.DB.Create(notification)

			// 创建操作日志
			database.DB.Create(&models.OperationLog{
				UserID:      ticket.CurrentUserID,
				Module:      "notification",
				Action:      "POST",
				Description: "收到工单超时通知",
			})

			// 如果严重超时，通知管理员
			if timeoutLevel == "severe" {
				var admins []models.User
				database.DB.Where("role = ?", "admin").Find(&admins)
				for _, admin := range admins {
					notification := &models.Notification{
						UserID:   admin.ID,
						Type:     "timeout",
						Title:    "严重超时工单提醒",
						Content:  fmt.Sprintf("工单 %s 已严重超时，请关注", ticket.JiraKey),
						TicketID: ticket.ID,
					}
					database.DB.Create(notification)

					// 创建操作日志
					database.DB.Create(&models.OperationLog{
						UserID:      admin.ID,
						Module:      "notification",
						Action:      "POST",
						Description: "收到严重超时工单通知",
					})
				}
			}
		}
	}

	return nil
}

// GetStatistics 获取统计数据
func (s *TicketService) GetStatistics() (*StatisticsResponse, error) {
	var stats StatisticsResponse

	// 人员维度统计
	var userStats []UserStatistic
	if err := database.DB.Model(&models.User{}).
		Select("id, name, role, ticket_num").
		Find(&userStats).Error; err != nil {
		return nil, err
	}

	// 填充每个用户的详细统计
	for i := range userStats {
		// 处理工单总数
		var totalTickets int64
		database.DB.Model(&models.TicketFlow{}).
			Where("to_user_id = ?", userStats[i].ID).
			Count(&totalTickets)
		userStats[i].TotalTickets = totalTickets

		// 平均处理时长
		var avgProcessTime sql.NullFloat64
		database.DB.Model(&models.TicketFlow{}).
			Where("to_user_id = ?", userStats[i].ID).
			Select("AVG(process_time)").
			Scan(&avgProcessTime)
		if avgProcessTime.Valid {
			userStats[i].AvgProcessTime = avgProcessTime.Float64
		} else {
			userStats[i].AvgProcessTime = 0
		}

		// 超时次数
		var timeoutCount int64
		database.DB.Model(&models.TicketFlow{}).
			Where("to_user_id = ? AND is_timeout = ?", userStats[i].ID, true).
			Count(&timeoutCount)
		userStats[i].TimeoutCount = timeoutCount
	}

	stats.UserStats = userStats

	// 工单维度统计
	var ticketStats []TicketStatistic
	if err := database.DB.Model(&models.Ticket{}).
		Select("id, jira_key, status, created_at").
		Find(&ticketStats).Error; err != nil {
		return nil, err
	}

	// 填充每个工单的详细统计
	for i := range ticketStats {
		// 总处理时长
		var totalDuration float64
		database.DB.Model(&models.TicketFlow{}).
			Where("ticket_id = ?", ticketStats[i].ID).
			Select("SUM(process_time)").
			Scan(&totalDuration)
		ticketStats[i].TotalDuration = totalDuration

		// 流转次数
		var flowCount int64
		database.DB.Model(&models.TicketFlow{}).
			Where("ticket_id = ?", ticketStats[i].ID).
			Count(&flowCount)
		ticketStats[i].FlowCount = flowCount
	}

	stats.TicketStats = ticketStats

	return &stats, nil
}

// 请求和响应结构体
type CreateTicketRequest struct {
	JiraKey       string `json:"jira_key" binding:"required"`
	JiraURL       string `json:"jira_url" binding:"required"`
	Description   string `json:"description"`
	Priority      string `json:"priority"`
	CurrentUserID uint   `json:"current_user_id" binding:"required"`
	CreatorID     uint   `json:"creator_id"`
}

type FlowTicketRequest struct {
	TicketID   uint   `json:"ticket_id" binding:"required"`
	FromUserID uint   `json:"from_user_id" binding:"required"`
	ToUserID   uint   `json:"to_user_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Status     string `json:"status"`
}

type StatisticsResponse struct {
	UserStats   []UserStatistic   `json:"user_stats"`
	TicketStats []TicketStatistic `json:"ticket_stats"`
}

type UserStatistic struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Role           string  `json:"role"`
	TicketNum      int     `json:"ticket_num"`
	TotalTickets   int64   `json:"total_tickets"`
	AvgProcessTime float64 `json:"avg_process_time"`
	TimeoutCount   int64   `json:"timeout_count"`
}

type TicketStatistic struct {
	ID            uint      `json:"id"`
	JiraKey       string    `json:"jira_key"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	TotalDuration float64   `json:"total_duration"`
	FlowCount     int64     `json:"flow_count"`
}

// DashboardData 数据看板数据
type DashboardData struct {
	TicketOverview   TicketOverview   `json:"ticket_overview"`
	TodayStatistics  TodayStatistics  `json:"today_statistics"`
	UserWorkload     []UserWorkload   `json:"user_workload"`
	TimeoutAlert     []TimeoutAlert   `json:"timeout_alert"`
	RecentActivities []RecentActivity `json:"recent_activities"`
}

// TicketOverview 工单概览
type TicketOverview struct {
	Total          int64 `json:"total"`
	Processing     int64 `json:"processing"`
	Retesting      int64 `json:"retesting"`
	Closed         int64 `json:"closed"`
	TimeoutCount   int64 `json:"timeout_count"`
	TotalTrend     int   `json:"total_trend"`      // 工单总数较昨日变化百分比
	ProcessingTrend int  `json:"processing_trend"` // 处理中较昨日变化百分比
	RetestingTrend int   `json:"retesting_trend"`  // 待复测较昨日变化百分比
	TimeoutTrend   int   `json:"timeout_trend"`    // 超时工单较昨日变化百分比
}

// TodayStatistics 今日统计
type TodayStatistics struct {
	NewTickets     int64 `json:"new_tickets"`
	ClosedTickets  int64 `json:"closed_tickets"`
	TimeoutTickets int64 `json:"timeout_tickets"`
	AvgProcessTime int64 `json:"avg_process_time"`
}

// UserWorkload 用户工作负载
type UserWorkload struct {
	UserID        uint   `json:"user_id"`
	UserName      string `json:"user_name"`
	TicketNum     int    `json:"ticket_num"`
	TimeoutCount  int64  `json:"timeout_count"`
	AvgHandleTime int64  `json:"avg_handle_time"`
}

// TimeoutAlert 超时预警
type TimeoutAlert struct {
	TicketID     uint      `json:"ticket_id"`
	JiraKey      string    `json:"jira_key"`
	TimeoutLevel string    `json:"timeout_level"`
	Handler      string    `json:"handler"`
	CreatedAt    time.Time `json:"created_at"`
}

// RecentActivity 最近活动
type RecentActivity struct {
	Type        string    `json:"type"`
	TicketKey   string    `json:"ticket_key"`
	Description string    `json:"description"`
	Operator    string    `json:"operator"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetDashboard 获取数据看板
func (s *TicketService) GetDashboard() (*DashboardData, error) {
	data := &DashboardData{}

	// 获取今日和昨日的日期字符串
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 1. 工单概览 - 今日数据
	overview := TicketOverview{}
	database.DB.Model(&models.Ticket{}).Count(&overview.Total)
	database.DB.Model(&models.Ticket{}).Where("status = ?", "processing").Count(&overview.Processing)
	database.DB.Model(&models.Ticket{}).Where("status = ?", "retesting").Count(&overview.Retesting)
	database.DB.Model(&models.Ticket{}).Where("status = ?", "closed").Count(&overview.Closed)
	database.DB.Model(&models.Ticket{}).Where("is_timeout = ?", true).Count(&overview.TimeoutCount)

	// 计算趋势 - 昨日数据
	var yesterdayTotal, yesterdayProcessing, yesterdayRetesting, yesterdayTimeout int64
	database.DB.Model(&models.Ticket{}).Where("DATE(created_at) <= ?", yesterday).Count(&yesterdayTotal)
	database.DB.Model(&models.Ticket{}).Where("status = ? AND DATE(updated_at) <= ?", "processing", yesterday).Count(&yesterdayProcessing)
	database.DB.Model(&models.Ticket{}).Where("status = ? AND DATE(updated_at) <= ?", "retesting", yesterday).Count(&yesterdayRetesting)
	database.DB.Model(&models.Ticket{}).Where("is_timeout = ? AND DATE(updated_at) <= ?", true, yesterday).Count(&yesterdayTimeout)

	// 计算变化百分比
	overview.TotalTrend = calculateTrend(overview.Total, yesterdayTotal)
	overview.ProcessingTrend = calculateTrend(overview.Processing, yesterdayProcessing)
	overview.RetestingTrend = calculateTrend(overview.Retesting, yesterdayRetesting)
	overview.TimeoutTrend = calculateTrend(overview.TimeoutCount, yesterdayTimeout)

	data.TicketOverview = overview

	// 2. 今日统计
	todayStats := TodayStatistics{}
	database.DB.Model(&models.Ticket{}).Where("DATE(created_at) = ?", today).Count(&todayStats.NewTickets)
	database.DB.Model(&models.Ticket{}).Where("status = ? AND DATE(updated_at) = ?", "closed", today).Count(&todayStats.ClosedTickets)
	database.DB.Model(&models.Ticket{}).Where("is_timeout = ? AND DATE(updated_at) = ?", true, today).Count(&todayStats.TimeoutTickets)

	// 今日平均处理时长
	var avgTime float64
	database.DB.Model(&models.TicketFlow{}).
		Where("DATE(created_at) = ?", today).
		Select("AVG(process_time)").
		Scan(&avgTime)
	todayStats.AvgProcessTime = int64(avgTime)
	data.TodayStatistics = todayStats

	// 3. 用户工作负载
	var users []models.User
	database.DB.Where("status = ?", 1).Find(&users)

	workloads := []UserWorkload{}
	for _, user := range users {
		var timeoutCount int64
		var avgHandleTime *float64

		database.DB.Model(&models.TicketFlow{}).
			Where("to_user_id = ? AND is_timeout = ?", user.ID, true).
			Count(&timeoutCount)

		database.DB.Model(&models.TicketFlow{}).
			Where("to_user_id = ?", user.ID).
			Select("AVG(process_time)").
			Scan(&avgHandleTime)

		var avgHandleTimeValue int64
		if avgHandleTime != nil {
			avgHandleTimeValue = int64(*avgHandleTime)
		}

		workloads = append(workloads, UserWorkload{
			UserID:        user.ID,
			UserName:      user.Name,
			TicketNum:     user.TicketNum,
			TimeoutCount:  timeoutCount,
			AvgHandleTime: avgHandleTimeValue,
		})
	}
	data.UserWorkload = workloads

	// 4. 超时预警（当前超时的工单）
	var timeoutTickets []models.Ticket
	database.DB.Where("is_timeout = ?", true).
		Preload("CurrentUser").
		Order("updated_at desc").
		Limit(10).
		Find(&timeoutTickets)

	alerts := []TimeoutAlert{}
	for _, ticket := range timeoutTickets {
		alerts = append(alerts, TimeoutAlert{
			TicketID:     ticket.ID,
			JiraKey:      ticket.JiraKey,
			TimeoutLevel: ticket.TimeoutLevel,
			Handler:      ticket.CurrentUser.Name,
			CreatedAt:    ticket.UpdatedAt,
		})
	}
	data.TimeoutAlert = alerts

	// 5. 最近活动（最近10条流转记录）
	var recentFlows []models.TicketFlow
	database.DB.Preload("Ticket").
		Preload("FromUser").
		Order("created_at desc").
		Limit(10).
		Find(&recentFlows)

	activities := []RecentActivity{}
	for _, flow := range recentFlows {
		activityType := "流转"
		if flow.Content == "创建工单" {
			activityType = "创建"
		} else if len(flow.Content) > 6 && flow.Content[:6] == "复测结论" {
			activityType = "复测"
		}

		activities = append(activities, RecentActivity{
			Type:        activityType,
			TicketKey:   flow.Ticket.JiraKey,
			Description: flow.Content,
			Operator:    flow.FromUser.Name,
			CreatedAt:   flow.CreatedAt,
		})
	}
	data.RecentActivities = activities

	return data, nil
}

// BatchAssignRequest 批量分配请求
type BatchAssignRequest struct {
	TicketIDs  []uint `json:"ticket_ids" binding:"required"`
	ToUserID   uint   `json:"to_user_id" binding:"required"`
	Content    string `json:"content"`
	OperatorID uint   `json:"operator_id" binding:"required"`
}

// BatchAssignTickets 批量分配工单
func (s *TicketService) BatchAssignTickets(req *BatchAssignRequest) (int, error) {
	successCount := 0

	for _, ticketID := range req.TicketIDs {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			var ticket models.Ticket
			if err := tx.First(&ticket, ticketID).Error; err != nil {
				return err
			}

			// 创建流转记录
			flow := &models.TicketFlow{
				TicketID:   ticketID,
				FromUserID: ticket.CurrentUserID,
				ToUserID:   req.ToUserID,
				Content:    req.Content,
				IsTimeout:  ticket.IsTimeout,
			}
			if err := tx.Create(flow).Error; err != nil {
				return err
			}

			// 更新工单
			if err := tx.Model(&ticket).Updates(map[string]any{
				"current_user_id": req.ToUserID,
				"is_timeout":      false,
				"timeout_level":   "",
			}).Error; err != nil {
				return err
			}

			// 更新用户待处理工单数
			tx.Model(&models.User{}).Where("id = ?", ticket.CurrentUserID).
				UpdateColumn("ticket_num", gorm.Expr("ticket_num - ?", 1))
			tx.Model(&models.User{}).Where("id = ?", req.ToUserID).
				UpdateColumn("ticket_num", gorm.Expr("ticket_num + ?", 1))

			return nil
		})

		if err == nil {
			successCount++
		}
	}

	return successCount, nil
}

// calculateTrend 计算变化百分比
func calculateTrend(today, yesterday int64) int {
	if yesterday == 0 {
		if today > 0 {
			return 100 // 从0增长视为100%
		}
		return 0
	}
	change := float64(today-yesterday) / float64(yesterday) * 100
	return int(change)
}

// BatchCloseRequest 批量关闭请求
type BatchCloseRequest struct {
	TicketIDs  []uint `json:"ticket_ids" binding:"required"`
	Conclusion string `json:"conclusion" binding:"required"`
	OperatorID uint   `json:"operator_id" binding:"required"`
}

// BatchCloseTickets 批量关闭工单
func (s *TicketService) BatchCloseTickets(req *BatchCloseRequest) (int, error) {
	successCount := 0

	for _, ticketID := range req.TicketIDs {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			var ticket models.Ticket
			if err := tx.First(&ticket, ticketID).Error; err != nil {
				return err
			}

			if ticket.Status == "closed" {
				return errors.New("工单已关闭")
			}

			// 创建流转记录
			flow := &models.TicketFlow{
				TicketID:   ticketID,
				FromUserID: ticket.CurrentUserID,
				ToUserID:   req.OperatorID,
				Content:    "批量关闭: " + req.Conclusion,
				IsTimeout:  ticket.IsTimeout,
			}
			if err := tx.Create(flow).Error; err != nil {
				return err
			}

			// 更新工单状态
			if err := tx.Model(&ticket).Update("status", "closed").Error; err != nil {
				return err
			}

			// 更新用户待处理工单数
			tx.Model(&models.User{}).Where("id = ?", ticket.CurrentUserID).
				UpdateColumn("ticket_num", gorm.Expr("ticket_num - ?", 1))

			return nil
		})

		if err == nil {
			successCount++
		}
	}

	return successCount, nil
}

// calculateTrend 计算变化百分比
