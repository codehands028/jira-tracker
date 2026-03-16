package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"jira-tracker-backend/config"
	"jira-tracker-backend/database"
	"jira-tracker-backend/models"

	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	taskMutex sync.RWMutex
	tasks     map[uint]*models.ExportTask // 内存中存储任务状态
}

func NewExportService() *ExportService {
	return &ExportService{
		tasks: make(map[uint]*models.ExportTask),
	}
}

// getStoragePath 获取文件存储路径，如果配置中没有设置则使用默认路径
func (s *ExportService) getStoragePath() string {
	if config.GlobalConfig != nil && config.GlobalConfig.Export.StoragePath != "" {
		return config.GlobalConfig.Export.StoragePath
	}
	return "/tmp" // 默认路径
}

// ensureStoragePath 确保存储路径存在
func (s *ExportService) ensureStoragePath() error {
	storagePath := s.getStoragePath()
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return fmt.Errorf("创建存储目录失败: %w", err)
	}
	return nil
}

// ExportRequest 导出请求
type ExportRequest struct {
	UserID        uint   `json:"user_id"`
	UserRole      string `json:"user_role"`
	StartDate     string `json:"start_date"`      // 开始日期 YYYY-MM-DD
	EndDate       string `json:"end_date"`        // 结束日期 YYYY-MM-DD
	Status        string `json:"status"`          // 工单状态
	Priority      string `json:"priority"`        // 优先级
	TimeDimension string `json:"time_dimension"`  // 时间维度 day/week/month
	Page          int    `json:"page"`            // 页码（分页导出）
	PageSize      int    `json:"page_size"`       // 每页数量
}

// ExportTicketData 导出的工单数据
type ExportTicketData struct {
	JiraKey        string `json:"jira_key"`
	JiraURL        string `json:"jira_url"`
	Description    string `json:"description"`
	TicketType     string `json:"ticket_type"`
	Priority       string `json:"priority"`
	Status         string `json:"status"`
	CurrentHandler string `json:"current_handler"`
	CreatedAt      string `json:"created_at"`
	FlowTime       string `json:"flow_time"`
	ClosedAt       string `json:"closed_at"`
	NodeStayTimes  string `json:"node_stay_times"` // 各节点停留时间
	ProcessTime    string `json:"process_time"`    // 总处理时长
	IsTimeout      bool   `json:"is_timeout"`
	TimeoutLevel   string `json:"timeout_level"`
}

// CreateExportTask 创建导出任务
func (s *ExportService) CreateExportTask(userID uint) (*models.ExportTask, error) {
	task := &models.ExportTask{
		UserID:    userID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// 保存到数据库
	if err := database.DB.Create(task).Error; err != nil {
		return nil, err
	}

	// 同时保存到内存
	s.taskMutex.Lock()
	s.tasks[task.ID] = task
	s.taskMutex.Unlock()

	return task, nil
}

// GetExportTask 从数据库获取导出任务
func (s *ExportService) GetExportTask(taskID uint) (*models.ExportTask, error) {
	// 先从内存获取
	s.taskMutex.RLock()
	if task, ok := s.tasks[taskID]; ok {
		s.taskMutex.RUnlock()
		return task, nil
	}
	s.taskMutex.RUnlock()

	// 从数据库获取
	var task models.ExportTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateExportTask 更新导出任务状态
func (s *ExportService) UpdateExportTask(task *models.ExportTask) error {
	// 更新数据库
	if err := database.DB.Save(task).Error; err != nil {
		return err
	}

	// 更新内存
	s.taskMutex.Lock()
	s.tasks[task.ID] = task
	s.taskMutex.Unlock()

	return nil
}

// ExportToExcel 导出工单数据到Excel
func (s *ExportService) ExportToExcel(req *ExportRequest, progressCallback func(int, int)) (string, error) {
	// 设置默认分页大小
	if req.PageSize <= 0 {
		req.PageSize = 1000 // 默认每页1000条
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	// 查询工单数据
	tickets, total, err := s.getTicketsForExport(req)
	if err != nil {
		return "", err
	}

	// 通知总记录数
	if progressCallback != nil {
		progressCallback(0, total)
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 设置工作表名称
	sheetName := "工单数据"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // 删除默认工作表

	// 设置表头
	headers := []string{
		"工单编号(Jira Key)", "工单链接", "工单描述", "工单类型", "优先级", "当前状态",
		"当前处理人", "创建时间", "流转时间", "关闭时间",
		"各节点停留时间", "总处理时长", "是否超时", "超时级别",
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 设置列宽
	columnWidths := map[string]float64{
		"A": 18, "B": 35, "C": 40, "D": 12, "E": 10, "F": 12,
		"G": 12, "H": 20, "I": 20, "J": 20,
		"K": 50, "L": 15, "M": 10, "N": 10,
	}
	for col, width := range columnWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// 填充数据
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
	})

	timeoutStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Color: "#FF0000"},
	})

	for i, ticket := range tickets {
		row := i + 2

		// 更新进度
		if progressCallback != nil && i%10 == 0 { // 每10条更新一次进度，减少性能开销
			progressCallback(i+1, total)
		}

		// 设置是否超时的样式
		rowStyle := dataStyle
		if ticket.IsTimeout {
			rowStyle = timeoutStyle
		}

		cells := []interface{}{
			ticket.JiraKey,
			ticket.JiraURL,
			ticket.Description,
			s.getTicketTypeText(ticket.TicketType),
			s.getPriorityText(ticket.Priority),
			s.getStatusText(ticket.Status),
			ticket.CurrentHandler,
			ticket.CreatedAt,
			ticket.FlowTime,
			ticket.ClosedAt,
			ticket.NodeStayTimes,
			ticket.ProcessTime,
			s.getBoolText(ticket.IsTimeout),
			ticket.TimeoutLevel,
		}

		for j, value := range cells {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheetName, cell, value)
			f.SetCellStyle(sheetName, cell, cell, rowStyle)
		}
	}

	// 生成文件名
	filename := fmt.Sprintf("工单导出_%s.xlsx", time.Now().Format("20060102150405"))

	// 确保存储目录存在
	if err := s.ensureStoragePath(); err != nil {
		return "", err
	}

	// 使用配置的存储路径
	filepath := s.getStoragePath() + "/" + filename

	// 保存文件
	if err := f.SaveAs(filepath); err != nil {
		return "", err
	}

	// 如果是分页导出，添加分页信息到文件名
	if total > req.PageSize {
		filename = fmt.Sprintf("工单导出_第%d页_%s.xlsx", req.Page, time.Now().Format("20060102150405"))
		filepath = s.getStoragePath() + "/" + filename
		if err := f.SaveAs(filepath); err != nil {
			return "", err
		}
	}

	return filepath, nil
}

// ExportToExcelAsync 异步导出工单数据
func (s *ExportService) ExportToExcelAsync(req *ExportRequest, taskID uint) {
	// 更新任务状态为处理中
	task, _ := s.GetExportTask(taskID)
	if task == nil {
		return
	}

	task.Status = "processing"
	task.UpdatedAt = time.Now()
	s.UpdateExportTask(task)

	// 执行导出，使用进度回调
	filepath, err := s.ExportToExcel(req, func(processed, total int) {
		// 更新进度
		task.TotalCount = total
		task.ProcessCount = processed
		task.UpdatedAt = time.Now()
		s.UpdateExportTask(task)
	})

	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		task.UpdatedAt = time.Now()
		s.UpdateExportTask(task)
		return
	}

	// 更新任务状态为完成
	task.Status = "completed"
	task.FilePath = filepath
	task.Filename = filepath[len(s.getStoragePath())+1:]
	task.TotalCount = task.ProcessCount // 确保总数正确
	now := time.Now()
	task.CompletedAt = &now
	task.UpdatedAt = now
	s.UpdateExportTask(task)
}

// getTicketsForExport 获取要导出的工单数据
func (s *ExportService) getTicketsForExport(req *ExportRequest) ([]ExportTicketData, int, error) {
	query := database.DB.Model(&models.Ticket{}).Preload("CurrentUser")

	// 权限过滤：非管理员只能导出自己的工单
	if req.UserRole != "admin" {
		query = query.Where("current_user_id = ? OR id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)",
			req.UserID, req.UserID)
	}

	// 时间范围过滤
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			query = query.Where("created_at <= ?", endTime.Add(24*time.Hour))
		}
	}

	// 状态过滤
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 优先级过滤
	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var tickets []models.Ticket
	if err := query.Order("created_at desc").Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	// 转换为导出数据格式
	result := make([]ExportTicketData, 0, len(tickets))
	for _, ticket := range tickets {
		data := ExportTicketData{
			JiraKey:        ticket.JiraKey,
			JiraURL:        ticket.JiraURL,
			Description:    ticket.Description,
			TicketType:     ticket.Type,
			Priority:       ticket.Priority,
			Status:         ticket.Status,
			CurrentHandler: ticket.CurrentUser.Name,
			CreatedAt:      ticket.CreatedAt.Format("2006-01-02 15:04:05"),
			IsTimeout:      ticket.IsTimeout,
			TimeoutLevel:   ticket.TimeoutLevel,
		}

		// 获取流转记录
		var flows []models.TicketFlow
		database.DB.Where("ticket_id = ?", ticket.ID).
			Preload("FromUser").
			Preload("ToUser").
			Order("created_at asc").
			Find(&flows)

		// 计算各节点停留时间
		nodeTimes := make([]string, 0)
		var totalDuration time.Duration
		var flowTimeStr, closedTimeStr string

		for i, flow := range flows {
			if i > 0 {
				stayDuration := time.Duration(flow.ProcessTime) * time.Second
				nodeTimes = append(nodeTimes, fmt.Sprintf("%s: %s",
					flows[i-1].ToUser.Name, s.formatDuration(stayDuration)))
				totalDuration += stayDuration
			}

			// 记录流转时间和关闭时间
			if flow.Content == "创建工单" && len(flows) > 1 {
				flowTimeStr = flow.CreatedAt.Format("2006-01-02 15:04:05")
			}
			if ticket.Status == "closed" && i == len(flows)-1 {
				closedTimeStr = flow.CreatedAt.Format("2006-01-02 15:04:05")
			}
		}

		data.FlowTime = flowTimeStr
		data.ClosedAt = closedTimeStr
		// 将节点时间列表转为JSON字符串
		if len(nodeTimes) > 0 {
			nodeTimesJSON, _ := json.Marshal(nodeTimes)
			data.NodeStayTimes = string(nodeTimesJSON)
		}
		data.ProcessTime = s.formatDuration(totalDuration)

		result = append(result, data)
	}

	return result, int(total), nil
}

// formatDuration 格式化时长
func (s *ExportService) formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours >= 24 {
		days := hours / 24
		remainHours := hours % 24
		if remainHours > 0 {
			return fmt.Sprintf("%d天%d小时", days, remainHours)
		}
		return fmt.Sprintf("%d天", days)
	}

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%d小时%d分钟", hours, minutes)
		}
		return fmt.Sprintf("%d小时", hours)
	}

	return fmt.Sprintf("%d分钟", minutes)
}

// getTicketTypeText 获取工单类型文本
func (s *ExportService) getTicketTypeText(ticketType string) string {
	mapping := map[string]string{
		"bug":         "缺陷",
		"feature":     "功能",
		"task":        "任务",
		"improvement": "改进",
		"":            "未分类",
	}
	if text, ok := mapping[ticketType]; ok {
		return text
	}
	return ticketType
}

// getPriorityText 获取优先级文本
func (s *ExportService) getPriorityText(priority string) string {
	mapping := map[string]string{
		"low":      "低",
		"medium":   "中",
		"high":     "高",
		"critical": "紧急",
	}
	if text, ok := mapping[priority]; ok {
		return text
	}
	return priority
}

// getStatusText 获取状态文本
func (s *ExportService) getStatusText(status string) string {
	mapping := map[string]string{
		"processing": "处理中",
		"retesting":  "待复测",
		"closed":     "已关闭",
	}
	if text, ok := mapping[status]; ok {
		return text
	}
	return status
}

// getBoolText 获取布尔值文本
func (s *ExportService) getBoolText(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

// GetPersonalDashboard 获取个人数据看板
func (s *ExportService) GetPersonalDashboard(userID uint, role string, startDate, endDate, status, priority, timeDimension string) (*PersonalDashboardData, error) {
	data := &PersonalDashboardData{}

	// 时间范围处理
	var startTime, endTime time.Time
	var err error

	if startDate != "" {
		startTime, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			startTime = time.Now().AddDate(0, 0, -30) // 默认最近30天
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -30)
	}

	if endDate != "" {
		endTime, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			endTime = time.Now()
		} else {
			endTime = endTime.Add(24 * time.Hour)
		}
	} else {
		endTime = time.Now()
	}

	// 默认时间维度为天
	if timeDimension == "" {
		timeDimension = "day"
	}

	// 1. 待处理工单数量（支持状态筛选，支持时间范围）
	var pendingCount int64
	pendingQuery := database.DB.Model(&models.Ticket{}).
		Where("current_user_id = ?", userID)
	if status != "" {
		pendingQuery = pendingQuery.Where("status = ?", status)
	} else {
		pendingQuery = pendingQuery.Where("status IN (?)", []string{"processing", "retesting"})
	}
	if priority != "" {
		pendingQuery = pendingQuery.Where("priority = ?", priority)
	}
	// 添加时间范围筛选
	if !startTime.IsZero() {
		pendingQuery = pendingQuery.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		pendingQuery = pendingQuery.Where("created_at <= ?", endTime)
	}
	pendingQuery.Count(&pendingCount)
	data.PendingTickets = pendingCount

	// 2. 已处理工单数量（已处理并关闭的工单）
	var processedCount int64
	processedQuery := database.DB.Model(&models.Ticket{}).
		Where("status = ?", "closed").
		Where("id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)", userID)
	if status != "" && status != "closed" {
		// 如果筛选的状态不是closed，则已处理工单为0
		processedCount = 0
	} else {
		if priority != "" {
			processedQuery = processedQuery.Where("priority = ?", priority)
		}
		// 添加时间范围筛选（按关闭时间）
		if !startTime.IsZero() {
			processedQuery = processedQuery.Where("updated_at >= ?", startTime)
		}
		if !endTime.IsZero() {
			processedQuery = processedQuery.Where("updated_at <= ?", endTime)
		}
		processedQuery.Count(&processedCount)
	}
	data.ProcessedTickets = processedCount

	// 3. 平均处理时长
	var avgProcessTime float64
	avgQuery := database.DB.Model(&models.TicketFlow{}).
		Where("to_user_id = ? AND process_time > 0", userID)
	if priority != "" {
		avgQuery = avgQuery.Joins("JOIN tickets ON tickets.id = ticket_flows.ticket_id").
			Where("tickets.priority = ?", priority)
	}
	avgQuery.Select("AVG(process_time)").Scan(&avgProcessTime)
	data.AvgProcessTime = int64(avgProcessTime)

	// 4. 超时工单数（当前用户待处理中已超时的工单）
	var timeoutCount int64
	timeoutQuery := database.DB.Model(&models.Ticket{}).
		Where("current_user_id = ? AND is_timeout = ?", userID, true).
		Where("status IN (?)", []string{"processing", "retesting"})
	if priority != "" {
		timeoutQuery = timeoutQuery.Where("priority = ?", priority)
	}
	// 添加时间范围筛选
	if !startTime.IsZero() {
		timeoutQuery = timeoutQuery.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		timeoutQuery = timeoutQuery.Where("created_at <= ?", endTime)
	}
	timeoutQuery.Count(&timeoutCount)
	data.TimeoutTickets = timeoutCount

	// 5. 工单处理趋势（按时间维度统计）
	trendData, err := s.getTrendData(userID, role, startTime, endTime, status, priority, timeDimension)
	if err != nil {
		return nil, err
	}
	data.TrendData = trendData

	// 6. 按状态分布
	statusDistribution := s.getStatusDistribution(userID, role, priority)
	data.StatusDistribution = statusDistribution

	// 7. 按优先级分布
	priorityDistribution := s.getPriorityDistribution(userID, role, status)
	data.PriorityDistribution = priorityDistribution

	// 8. 按类型分布
	typeDistribution := s.getTypeDistribution(userID, role)
	data.TypeDistribution = typeDistribution

	return data, nil
}

// TrendData 趋势数据
type TrendData struct {
	Date      string `json:"date"`
	Count     int64  `json:"count"`
	Timeout   int64  `json:"timeout"`
	Processed int64  `json:"processed"`
}

// StatusDistribution 状态分布
type StatusDistribution struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// PriorityDistribution 优先级分布
type PriorityDistribution struct {
	Priority string `json:"priority"`
	Count    int64  `json:"count"`
}

// TypeDistribution 类型分布
type TypeDistribution struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// PersonalDashboardData 个人数据看板数据
type PersonalDashboardData struct {
	PendingTickets       int64                `json:"pending_tickets"`        // 待处理工单
	ProcessedTickets     int64                `json:"processed_tickets"`      // 已处理工单
	AvgProcessTime       int64                `json:"avg_process_time"`       // 平均处理时长(秒)
	TimeoutTickets       int64                `json:"timeout_tickets"`        // 超时工单数
	TrendData            []TrendData          `json:"trend_data"`             // 趋势数据
	StatusDistribution   []StatusDistribution `json:"status_distribution"`    // 状态分布
	PriorityDistribution []PriorityDistribution `json:"priority_distribution"` // 优先级分布
	TypeDistribution     []TypeDistribution   `json:"type_distribution"`      // 类型分布
}

// getTrendData 获取趋势数据（支持日/周/月维度）
func (s *ExportService) getTrendData(userID uint, role string, startTime, endTime time.Time, status, priority, timeDimension string) ([]TrendData, error) {
	var result []TrendData

	// 根据时间维度生成时间序列
	var intervals []time.Time
	switch timeDimension {
	case "week":
		// 按周统计
		for t := startTime; t.Before(endTime); t = t.AddDate(0, 0, 7) {
			intervals = append(intervals, t)
		}
	case "month":
		// 按月统计
		for t := startTime; t.Before(endTime); t = t.AddDate(0, 1, 0) {
			intervals = append(intervals, t)
		}
	default:
		// 按天统计
		days := int(endTime.Sub(startTime).Hours() / 24)
		if days > 90 {
			days = 90 // 最多90天
		}
		for i := 0; i <= days; i++ {
			intervals = append(intervals, startTime.AddDate(0, 0, i))
		}
	}

	for i, t := range intervals {
		var nextTime time.Time
		if i+1 < len(intervals) {
			nextTime = intervals[i+1]
		} else {
			switch timeDimension {
			case "week":
				nextTime = t.AddDate(0, 0, 7)
			case "month":
				nextTime = t.AddDate(0, 1, 0)
			default:
				nextTime = t.AddDate(0, 0, 1)
			}
		}
		if nextTime.After(endTime) {
			nextTime = endTime
		}

		var dateLabel string
		switch timeDimension {
		case "week":
			dateLabel = t.Format("2006-01-02") + " 周"
		case "month":
			dateLabel = t.Format("2006-01")
		default:
			dateLabel = t.Format("2006-01-02")
		}

		var count, timeout, processed int64

		// 当日新增工单
		query := database.DB.Model(&models.Ticket{}).Where("created_at >= ? AND created_at < ?", t, nextTime)
		if role != "admin" {
			query = query.Where("current_user_id = ? OR id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)",
				userID, userID)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if priority != "" {
			query = query.Where("priority = ?", priority)
		}
		query.Count(&count)

		// 当日超时工单
		timeoutQuery := database.DB.Model(&models.TicketFlow{}).
			Where("created_at >= ? AND created_at < ? AND is_timeout = ?", t, nextTime, true)
		if role != "admin" {
			timeoutQuery = timeoutQuery.Where("to_user_id = ?", userID)
		}
		timeoutQuery.Count(&timeout)

		// 当日处理工单
		processedQuery := database.DB.Model(&models.TicketFlow{}).
			Where("created_at >= ? AND created_at < ?", t, nextTime)
		if role != "admin" {
			processedQuery = processedQuery.Where("to_user_id = ?", userID)
		}
		processedQuery.Count(&processed)

		result = append(result, TrendData{
			Date:      dateLabel,
			Count:     count,
			Timeout:   timeout,
			Processed: processed,
		})
	}

	return result, nil
}

// getStatusDistribution 获取状态分布
func (s *ExportService) getStatusDistribution(userID uint, role string, priority string) []StatusDistribution {
	var result []StatusDistribution

	statuses := []string{"processing", "retesting", "closed"}
	statusNames := map[string]string{
		"processing": "处理中",
		"retesting":  "待复测",
		"closed":     "已关闭",
	}

	for _, status := range statuses {
		var count int64
		query := database.DB.Model(&models.Ticket{}).Where("status = ?", status)
		if role != "admin" {
			query = query.Where("current_user_id = ? OR id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)",
				userID, userID)
		}
		if priority != "" {
			query = query.Where("priority = ?", priority)
		}
		query.Count(&count)

		if count > 0 {
			result = append(result, StatusDistribution{
				Status: statusNames[status],
				Count:  count,
			})
		}
	}

	return result
}

// getPriorityDistribution 获取优先级分布
func (s *ExportService) getPriorityDistribution(userID uint, role string, status string) []PriorityDistribution {
	var result []PriorityDistribution

	priorities := []string{"low", "medium", "high", "critical"}
	priorityNames := map[string]string{
		"low":      "低",
		"medium":   "中",
		"high":     "高",
		"critical": "紧急",
	}

	for _, priority := range priorities {
		var count int64
		query := database.DB.Model(&models.Ticket{}).Where("priority = ?", priority)
		if role != "admin" {
			query = query.Where("current_user_id = ? OR id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)",
				userID, userID)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		query.Count(&count)

		if count > 0 {
			result = append(result, PriorityDistribution{
				Priority: priorityNames[priority],
				Count:    count,
			})
		}
	}

	return result
}

// getTypeDistribution 获取类型分布
func (s *ExportService) getTypeDistribution(userID uint, role string) []TypeDistribution {
	var result []TypeDistribution

	types := []string{"bug", "feature", "task", "improvement", ""}
	typeNames := map[string]string{
		"bug":         "缺陷",
		"feature":     "功能",
		"task":        "任务",
		"improvement": "改进",
		"":            "未分类",
	}

	for _, ticketType := range types {
		var count int64
		query := database.DB.Model(&models.Ticket{}).Where("type = ?", ticketType)
		if role != "admin" {
			query = query.Where("current_user_id = ? OR id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)",
				userID, userID)
		}
		query.Count(&count)

		if count > 0 {
			result = append(result, TypeDistribution{
				Type:  typeNames[ticketType],
				Count: count,
			})
		}
	}

	return result
}

// GetDashboardTickets 获取看板详情工单列表（点击图表后查看）
func (s *ExportService) GetDashboardTickets(userID uint, role string, date, status, priority, ticketType string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	query := database.DB.Model(&models.Ticket{}).Preload("CurrentUser")

	// 权限过滤
	if role != "admin" {
		query = query.Where("current_user_id = ? OR id IN (SELECT ticket_id FROM ticket_flows WHERE to_user_id = ?)",
			userID, userID)
	}

	// 日期筛选
	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			query = query.Where("created_at >= ? AND created_at < ?", t, t.AddDate(0, 0, 1))
		}
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 优先级筛选
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// 类型筛选
	if ticketType != "" {
		query = query.Where("type = ?", ticketType)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页
	offset := (page - 1) * pageSize
	var tickets []models.Ticket
	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	// 转换为前端需要的格式
	result := make([]map[string]interface{}, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, map[string]interface{}{
			"id":             ticket.ID,
			"jira_key":       ticket.JiraKey,
			"jira_url":       ticket.JiraURL,
			"description":    ticket.Description,
			"type":           ticket.Type,
			"priority":       ticket.Priority,
			"status":         ticket.Status,
			"current_user":   ticket.CurrentUser,
			"is_timeout":     ticket.IsTimeout,
			"timeout_level":  ticket.TimeoutLevel,
			"created_at":     ticket.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return result, total, nil
}

// GetUserConfig 获取用户配置
func (s *ExportService) GetUserConfig(userID uint) (*models.UserConfig, error) {
	var config models.UserConfig
	result := database.DB.Where("user_id = ?", userID).First(&config)
	if result.Error != nil {
		// 如果不存在，创建默认配置
		config = models.UserConfig{
			UserID:          userID,
			DashboardConfig: `{"showPending":true,"showProcessed":true,"showTimeout":true,"showAvgTime":true,"showTrend":true,"showStatus":true,"showPriority":true,"showType":true}`,
		}
		database.DB.Create(&config)
	}
	return &config, nil
}

// UpdateUserConfig 更新用户配置
func (s *ExportService) UpdateUserConfig(userID uint, configStr string) error {
	var config models.UserConfig
	result := database.DB.Where("user_id = ?", userID).First(&config)
	if result.Error != nil {
		// 创建新配置
		config = models.UserConfig{
			UserID:          userID,
			DashboardConfig: configStr,
		}
		return database.DB.Create(&config).Error
	}
	config.DashboardConfig = configStr
	return database.DB.Save(&config).Error
}
