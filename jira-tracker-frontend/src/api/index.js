import request from '@/utils/request'

// 认证相关
export const sendCode = (phone) => {
  return request({
    url: '/auth/send-code',
    method: 'post',
    data: { phone }
  })
}

export const login = (phone, code) => {
  return request({
    url: '/auth/login',
    method: 'post',
    data: { phone, code }
  })
}

// 用户管理
export const getUsers = (params) => {
  return request({
    url: '/users',
    method: 'get',
    params
  })
}

export const createUser = (data) => {
  return request({
    url: '/users',
    method: 'post',
    data
  })
}

export const updateUser = (id, data) => {
  return request({
    url: `/users/${id}`,
    method: 'put',
    data
  })
}

export const deleteUser = (id) => {
  return request({
    url: `/users/${id}`,
    method: 'delete'
  })
}

// 工单管理
export const getTickets = (params) => {
  return request({
    url: '/tickets',
    method: 'get',
    params
  })
}

export const getTicketDetail = (id) => {
  return request({
    url: `/tickets/${id}`,
    method: 'get'
  })
}

export const createTicket = (data) => {
  return request({
    url: '/tickets',
    method: 'post',
    data
  })
}

export const flowTicket = (id, data) => {
  return request({
    url: `/tickets/${id}/flow`,
    method: 'post',
    data
  })
}

export const retestTicket = (id, data) => {
  return request({
    url: `/tickets/${id}/retest`,
    method: 'post',
    data
  })
}

export const closeTicket = (id, data) => {
  return request({
    url: `/tickets/${id}/close`,
    method: 'post',
    data
  })
}

// 批量操作
export const batchAssignTickets = (data) => {
  return request({
    url: '/tickets/batch/assign',
    method: 'post',
    data
  })
}

export const batchFlowTickets = (data) => {
  return request({
    url: '/tickets/batch/flow',
    method: 'post',
    data
  })
}

export const batchCloseTickets = (data) => {
  return request({
    url: '/tickets/batch/close',
    method: 'post',
    data
  })
}

// 统计数据
export const getStatistics = () => {
  return request({
    url: '/statistics',
    method: 'get'
  })
}

export const getDashboard = () => {
  return request({
    url: '/statistics/dashboard',
    method: 'get'
  })
}

// 通知
export const getNotifications = (params) => {
  return request({
    url: '/notifications',
    method: 'get',
    params
  })
}

export const markNotificationAsRead = (id) => {
  return request({
    url: `/notifications/${id}/read`,
    method: 'put'
  })
}

// 个人信息
export const getProfile = () => {
  return request({
    url: '/profile',
    method: 'get'
  })
}

export const updateProfile = (data) => {
  return request({
    url: '/profile',
    method: 'put',
    data
  })
}

// 操作日志
export const getOperationLogs = (params) => {
  return request({
    url: '/logs',
    method: 'get',
    params
  })
}

export const getMyLogs = (params) => {
  return request({
    url: '/logs/my',
    method: 'get',
    params
  })
}

// 超时规则管理
export const getTimeoutRules = () => {
  return request({
    url: '/timeout-rules',
    method: 'get'
  })
}

export const getActiveTimeoutRule = () => {
  return request({
    url: '/timeout-rules/active',
    method: 'get'
  })
}

export const createTimeoutRule = (data) => {
  return request({
    url: '/timeout-rules',
    method: 'post',
    data
  })
}

export const updateTimeoutRule = (id, data) => {
  return request({
    url: `/timeout-rules/${id}`,
    method: 'put',
    data
  })
}

export const setActiveTimeoutRule = (id) => {
  return request({
    url: `/timeout-rules/${id}/active`,
    method: 'put'
  })
}

export const setInactiveTimeoutRule = (id) => {
  return request({
    url: `/timeout-rules/${id}/inactive`,
    method: 'put'
  })
}

export const deleteTimeoutRule = (id) => {
  return request({
    url: `/timeout-rules/${id}`,
    method: 'delete'
  })
}

// 时限规则管理
export const getSLARules = () => {
  return request({
    url: '/sla-rules',
    method: 'get'
  })
}

export const getSLARule = (id) => {
  return request({
    url: `/sla-rules/${id}`,
    method: 'get'
  })
}

export const createSLARule = (data) => {
  return request({
    url: '/sla-rules',
    method: 'post',
    data
  })
}

export const updateSLARule = (id, data) => {
  return request({
    url: `/sla-rules/${id}`,
    method: 'put',
    data
  })
}

export const toggleSLARule = (id, isActive) => {
  return request({
    url: `/sla-rules/${id}/toggle`,
    method: 'put',
    data: { is_active: isActive }
  })
}

export const deleteSLARule = (id) => {
  return request({
    url: `/sla-rules/${id}`,
    method: 'delete'
  })
}

// 批量操作日志
export const getBatchOperationLogs = (params) => {
  return request({
    url: '/batch-logs',
    method: 'get',
    params
  })
}

// 个人数据看板
export const getPersonalDashboard = (params) => {
  return request({
    url: '/statistics/personal-dashboard',
    method: 'get',
    params
  })
}

// 导出工单Excel
export const exportTicketsExcel = (data) => {
  return request({
    url: '/export/excel',
    method: 'post',
    data
  })
}

// 异步导出工单Excel
export const exportTicketsExcelAsync = (data) => {
  return request({
    url: '/export/excel/async',
    method: 'post',
    data
  })
}

// 下载Excel文件
export const downloadExcelFile = (filename) => {
  return request({
    url: '/export/download',
    method: 'get',
    params: { filename },
    responseType: 'blob'
  })
}

// 获取导出任务状态
export const getExportTaskStatus = (taskId) => {
  return request({
    url: `/export/task/${taskId}`,
    method: 'get'
  })
}

// 获取用户配置
export const getUserConfig = () => {
  return request({
    url: '/user-config',
    method: 'get'
  })
}

// 更新用户配置
export const updateUserConfig = (data) => {
  return request({
    url: '/user-config',
    method: 'put',
    data
  })
}

// 获取看板详情工单列表
export const getDashboardTickets = (params) => {
  return request({
    url: '/dashboard/tickets',
    method: 'get',
    params
  })
}
