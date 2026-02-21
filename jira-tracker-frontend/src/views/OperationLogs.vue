<template>
  <div class="logs-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>操作日志</h1>
        <p>查看系统操作记录</p>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-card">
      <div class="filter-row">
        <el-select v-model="queryParams.user_id" placeholder="选择用户" clearable filterable class="filter-item">
          <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
        </el-select>
        
        <el-select v-model="queryParams.module" placeholder="选择模块" clearable class="filter-item">
          <el-option label="认证" value="auth" />
          <el-option label="用户" value="user" />
          <el-option label="工单" value="ticket" />
          <el-option label="统计" value="statistics" />
        </el-select>
        
        <el-select v-model="queryParams.action" placeholder="选择操作" clearable class="filter-item">
          <el-option label="创建" value="POST" />
          <el-option label="更新" value="PUT" />
          <el-option label="删除" value="DELETE" />
          <el-option label="查询" value="GET" />
        </el-select>
        
        <el-date-picker
          v-model="timeRange"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DD HH:mm:ss"
          :shortcuts="shortcuts"
          class="date-picker"
        />
        
        <el-button type="primary" @click="fetchLogs">查询</el-button>
        <el-button @click="resetQuery">重置</el-button>
      </div>
    </div>

    <!-- 日志列表 -->
    <div class="logs-list">
      <div
        v-for="log in logs"
        :key="log.id"
        class="log-item"
      >
        <div class="log-icon" :class="getActionClass(log.action)">
          <el-icon :size="20">
            <component :is="getActionIcon(log.action)" />
          </el-icon>
        </div>
        
        <div class="log-content">
          <div class="log-header">
            <span class="user-name">{{ log.user?.name || '未知用户' }}</span>
            <el-tag :type="getModuleType(log.module)" size="small">
              {{ getModuleLabel(log.module) }}
            </el-tag>
            <el-tag :type="getActionType(log.action)" size="small" effect="dark">
              {{ getActionLabel(log.action) }}
            </el-tag>
          </div>
          
          <p class="log-desc">{{ getOperationDescription(log) }}</p>
          
          <div class="log-footer">
            <span class="log-time">
              <el-icon><Clock /></el-icon>
              {{ formatTime(log.created_at) }}
            </span>
          </div>
        </div>
      </div>
      
      <el-empty v-if="logs.length === 0 && !loading" description="暂无日志记录" />
    </div>

    <!-- 分页 -->
    <div class="pagination-container" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchLogs"
        @current-change="fetchLogs"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getOperationLogs, getUsers } from '@/api'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

const loading = ref(false)
const logs = ref([])
const users = ref([])
const timeRange = ref([])

const queryParams = reactive({
  user_id: '',
  module: '',
  action: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const shortcuts = [
  {
    text: '今天',
    value: () => {
      const start = dayjs().startOf('day').format('YYYY-MM-DD HH:mm:ss')
      const end = dayjs().endOf('day').format('YYYY-MM-DD HH:mm:ss')
      return [start, end]
    }
  },
  {
    text: '最近一周',
    value: () => {
      const end = dayjs().format('YYYY-MM-DD HH:mm:ss')
      const start = dayjs().subtract(7, 'day').format('YYYY-MM-DD HH:mm:ss')
      return [start, end]
    }
  },
  {
    text: '最近一个月',
    value: () => {
      const end = dayjs().format('YYYY-MM-DD HH:mm:ss')
      const start = dayjs().subtract(1, 'month').format('YYYY-MM-DD HH:mm:ss')
      return [start, end]
    }
  }
]

const fetchUsers = async () => {
  try {
    const data = await getUsers({ page: 1, page_size: 1000 })
    users.value = data.list
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

const fetchLogs = async () => {
  try {
    loading.value = true
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }

    if (queryParams.user_id) params.user_id = queryParams.user_id
    if (queryParams.module) params.module = queryParams.module
    if (queryParams.action) params.action = queryParams.action
    if (timeRange.value && timeRange.value.length === 2) {
      params.start_time = timeRange.value[0]
      params.end_time = timeRange.value[1]
    }

    const data = await getOperationLogs(params)
    logs.value = data.list
    pagination.total = data.total
  } catch (error) {
    console.error('获取操作日志失败:', error)
  } finally {
    loading.value = false
  }
}

const resetQuery = () => {
  queryParams.user_id = ''
  queryParams.module = ''
  queryParams.action = ''
  timeRange.value = []
  pagination.page = 1
  fetchLogs()
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const getModuleLabel = (module) => {
  const map = {
    auth: '认证',
    user: '用户',
    ticket: '工单',
    notification: '通知',
    statistics: '统计',
    operation_log: '操作日志',
    timeout_rule: '超时规则'
  }
  return map[module] || module
}

const getModuleType = (module) => {
  const map = {
    auth: 'danger',
    user: 'warning',
    ticket: 'primary',
    notification: 'success',
    statistics: 'info',
    operation_log: 'info',
    timeout_rule: 'warning'
  }
  return map[module] || ''
}

const getActionType = (action) => {
  const map = {
    POST: 'success',
    PUT: 'warning',
    DELETE: 'danger',
    GET: 'info'
  }
  return map[action] || ''
}

const getActionLabel = (action) => {
  const map = {
    POST: '创建',
    PUT: '更新',
    DELETE: '删除',
    GET: '查询'
  }
  return map[action] || action
}

const getOperationDescription = (log) => {
  // 如果description为空、undefined或空字符串，返回默认描述
  if (!log.description || log.description.trim() === '') {
    const moduleLabel = getModuleLabel(log.module)
    return `${getActionLabel(log.action)}${moduleLabel}`
  }
  
  // 如果description不是API路径，直接返回
  if (!log.description.startsWith('/api/')) {
    return log.description
  }
  
  // 解析API路径，生成友好的描述
  const path = log.description
  const actionLabel = getActionLabel(log.action)
  
  // 根据不同的API路径生成不同的描述
  if (path.includes('/logs')) {
    return `查看操作日志`
  } else if (path.includes('/users')) {
    return `${actionLabel}用户信息`
  } else if (path.includes('/tickets')) {
    return `${actionLabel}工单`
  } else if (path.includes('/tickets/') && path.includes('/flows')) {
    return `${actionLabel}工单流转记录`
  } else if (path.includes('/notifications')) {
    return `${actionLabel}通知`
  } else if (path.includes('/statistics')) {
    return `查看统计数据`
  } else if (path.includes('/auth')) {
    return `用户${actionLabel === '查询' ? '登录' : actionLabel}操作`
  } else if (path.includes('/timeout-rules')) {
    return `${actionLabel}超时规则`
  } else {
    return `${actionLabel}操作`
  }
}

const getActionIcon = (action) => {
  const map = {
    POST: 'Plus',
    PUT: 'Edit',
    DELETE: 'Delete',
    GET: 'Search'
  }
  return map[action] || 'Document'
}

const getActionClass = (action) => {
  const map = {
    POST: 'success',
    PUT: 'warning',
    DELETE: 'danger',
    GET: 'info'
  }
  return map[action] || ''
}

onMounted(() => {
  fetchUsers()
  fetchLogs()
})
</script>

<style scoped>
.logs-container {
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
}

.header-left h1 {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.header-left p {
  color: var(--text-tertiary);
  font-size: 14px;
}

.filter-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.filter-item {
  width: 140px;
}

.date-picker {
  width: 340px;
}

.logs-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-item {
  background: white;
  border-radius: 12px;
  padding: 16px 20px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  display: flex;
  gap: 16px;
  transition: all var(--transition-fast);
}

.log-item:hover {
  box-shadow: var(--shadow-md);
}

.log-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.log-icon.success {
  background: rgba(72, 187, 120, 0.1);
  color: var(--success-color);
}

.log-icon.warning {
  background: rgba(237, 137, 54, 0.1);
  color: var(--warning-color);
}

.log-icon.danger {
  background: rgba(245, 101, 101, 0.1);
  color: var(--danger-color);
}

.log-icon.info {
  background: rgba(66, 153, 225, 0.1);
  color: var(--info-color);
}

.log-content {
  flex: 1;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.user-name {
  font-weight: 600;
  color: var(--text-primary);
}

.log-desc {
  color: var(--text-secondary);
  margin-bottom: 12px;
  font-size: 14px;
}

.log-footer {
  display: flex;
  gap: 20px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.log-ip,
.log-time {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  background: white;
  padding: 16px 20px;
  border-radius: 12px;
}

@media (max-width: 768px) {
  .filter-row {
    flex-direction: column;
  }
  
  .filter-item,
  .date-picker {
    width: 100%;
  }
}
</style>
