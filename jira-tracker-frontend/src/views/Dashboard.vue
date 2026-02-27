<template>
  <div class="dashboard-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>数据看板</h1>
        <p>实时监控工单处理状态</p>
      </div>
      <div class="header-right">
        <el-button @click="fetchDashboard" :icon="'Refresh'" circle />
        <span class="update-time">最后更新: {{ lastUpdateTime }}</span>
      </div>
    </div>

    <!-- 工单概览卡片 -->
    <div class="overview-cards">
      <div class="overview-card" v-for="(item, index) in overviewCards" :key="index">
        <div class="card-icon" :style="{ background: item.gradient }">
          <el-icon :size="28">
            <component :is="item.icon" />
          </el-icon>
        </div>
        <div class="card-content">
          <div class="card-value">{{ item.value }}</div>
          <div class="card-label">{{ item.label }}</div>
        </div>
        <div class="card-footer">
          <span :class="['trend', item.trend > 0 ? 'up' : 'down']">
            <el-icon><component :is="item.trend > 0 ? 'Top' : 'Bottom'" /></el-icon>
            {{ Math.abs(item.trend) }}%
          </span>
          <span class="trend-label">较昨日</span>
        </div>
      </div>
    </div>

    <!-- 今日统计 -->
    <div class="today-stats">
      <div class="stats-header">
        <h3>今日统计</h3>
        <span class="stats-date">{{ todayDate }}</span>
      </div>
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-icon new">
            <el-icon><Plus /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ todayStats.new_tickets }}</div>
            <div class="stat-label">新增工单</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon closed">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ todayStats.closed_tickets }}</div>
            <div class="stat-label">关闭工单</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon timeout">
            <el-icon><Warning /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value danger">{{ todayStats.timeout_tickets }}</div>
            <div class="stat-label">超时工单</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon time">
            <el-icon><Clock /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ formatDuration(todayStats.avg_process_time) }}</div>
            <div class="stat-label">平均处理时长</div>
          </div>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <!-- 用户工作负载 -->
      <div class="content-card workload-card">
        <div class="card-header">
          <h3>用户工作负载</h3>
          <el-tag size="small">实时</el-tag>
        </div>
        <div class="card-body">
          <div class="workload-list">
            <div
              v-for="user in userWorkload"
              :key="user.user_id"
              class="workload-item"
            >
              <div class="user-info">
                <el-avatar :size="40" class="user-avatar">
                  {{ user.user_name?.charAt(0) }}
                </el-avatar>
                <div class="user-detail">
                  <div class="user-name">{{ user.user_name }}</div>
                  <div class="user-stats">
                    <span class="stat">
                      <el-icon><Document /></el-icon>
                      {{ user.ticket_num }} 待处理
                    </span>
                    <span v-if="user.timeout_count > 0" class="stat timeout">
                      <el-icon><Warning /></el-icon>
                      {{ user.timeout_count }} 超时
                    </span>
                  </div>
                </div>
              </div>
              <div class="workload-bar">
                <div
                  class="bar-fill"
                  :style="{ width: getWorkloadPercent(user.ticket_num) + '%' }"
                  :class="{ warning: user.ticket_num > 5, danger: user.ticket_num > 8 }"
                ></div>
              </div>
            </div>
          </div>
          <el-empty v-if="userWorkload.length === 0" description="暂无数据" />
        </div>
      </div>

      <!-- 超时预警 -->
      <div class="content-card timeout-card">
        <div class="card-header">
          <h3>超时预警</h3>
          <el-badge :value="timeoutAlert.length" type="danger" />
        </div>
        <div class="card-body">
          <div class="timeout-list">
            <div
              v-for="alert in timeoutAlert"
              :key="alert.ticket_id"
              class="timeout-item"
              @click="goToTicket(alert.ticket_id)"
            >
              <div class="timeout-header">
                <router-link :to="`/tickets/${alert.ticket_id}`" class="ticket-link">
                  {{ alert.jira_key }}
                </router-link>
                <el-tag
                  :type="alert.timeout_level === 'severe' ? 'danger' : 'warning'"
                  size="small"
                  effect="dark"
                >
                  {{ alert.timeout_level === 'severe' ? '严重' : '普通' }}
                </el-tag>
              </div>
              <div class="timeout-meta">
                <span class="handler">
                  <el-icon><User /></el-icon>
                  {{ alert.handler }}
                </span>
                <span class="time">
                  <el-icon><Clock /></el-icon>
                  {{ formatTime(alert.created_at) }}
                </span>
              </div>
            </div>
          </div>
          <el-empty v-if="timeoutAlert.length === 0" description="暂无超时工单" />
        </div>
      </div>
    </div>

    <!-- 最近活动 -->
    <div class="content-card activities-card">
      <div class="card-header">
        <h3>最近活动</h3>
      </div>
      <div class="card-body">
        <div class="activity-timeline">
          <div
            v-for="(activity, index) in recentActivities"
            :key="index"
            class="activity-item"
          >
            <div class="activity-dot" :class="getActivityClass(activity.type)"></div>
            <div class="activity-content">
              <div class="activity-header">
                <el-tag
                  :type="getActivityType(activity.type)"
                  size="small"
                  effect="plain"
                >
                  {{ activity.type }}
                </el-tag>
                <span class="activity-ticket">{{ activity.ticket_key }}</span>
              </div>
              <div class="activity-desc">{{ activity.description }}</div>
              <div class="activity-footer">
                <span class="operator">
                  <el-icon><User /></el-icon>
                  {{ activity.operator }}
                </span>
                <span class="time">{{ formatTime(activity.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
        <el-empty v-if="recentActivities.length === 0" description="暂无活动记录" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard } from '@/api'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'

dayjs.extend(duration)

const router = useRouter()
const loading = ref(false)
const lastUpdateTime = ref('')

const overview = reactive({
  total: 0,
  processing: 0,
  retesting: 0,
  timeout_count: 0
})

const todayStats = reactive({
  new_tickets: 0,
  closed_tickets: 0,
  timeout_tickets: 0,
  avg_process_time: 0
})

const userWorkload = ref([])
const timeoutAlert = ref([])
const recentActivities = ref([])

const todayDate = computed(() => dayjs().format('YYYY年MM月DD日'))

const overviewCards = computed(() => [
  {
    icon: 'Document',
    label: '工单总数',
    value: overview.total,
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    trend: 12
  },
  {
    icon: 'Loading',
    label: '处理中',
    value: overview.processing,
    gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
    trend: -5
  },
  {
    icon: 'CircleCheck',
    label: '待复测',
    value: overview.retesting,
    gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    trend: 8
  },
  {
    icon: 'Warning',
    label: '超时工单',
    value: overview.timeout_count,
    gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    trend: -2
  }
])

const fetchDashboard = async () => {
  try {
    loading.value = true
    const data = await getDashboard()
    
    Object.assign(overview, data.ticket_overview)
    Object.assign(todayStats, data.today_statistics)
    userWorkload.value = data.user_workload || []
    timeoutAlert.value = data.timeout_alert || []
    recentActivities.value = data.recent_activities || []
    
    lastUpdateTime.value = dayjs().format('HH:mm:ss')
  } catch (error) {
    console.error('获取数据看板失败:', error)
  } finally {
    loading.value = false
  }
}

const goToTicket = (id) => {
  router.push(`/tickets/${id}`)
}

const formatTime = (time) => {
  return dayjs(time).format('MM-DD HH:mm')
}

const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const dur = dayjs.duration(seconds, 'seconds')
  const days = Math.floor(dur.asDays())
  const hours = dur.hours()
  const minutes = dur.minutes()

  let result = ''
  if (days > 0) result += `${days}天`
  if (hours > 0) result += `${hours}小时`
  if (minutes > 0 && days === 0) result += `${minutes}分钟`
  if (!result) result = '不到1分钟'

  return result
}

const getWorkloadPercent = (num) => {
  return Math.min(num * 10, 100)
}

const getActivityType = (type) => {
  const map = {
    '创建': 'success',
    '流转': 'primary',
    '复测': 'warning',
    '关闭': 'info'
  }
  return map[type] || ''
}

const getActivityClass = (type) => {
  const map = {
    '创建': 'success',
    '流转': 'primary',
    '复测': 'warning',
    '关闭': 'info'
  }
  return map[type] || 'primary'
}

let timer = null

onMounted(() => {
  fetchDashboard()
  timer = setInterval(fetchDashboard, 30000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.dashboard-container {
  padding: 0;
}

/* 页面标题 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.update-time {
  color: var(--text-tertiary);
  font-size: 13px;
}

/* 概览卡片 */
.overview-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.overview-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 24px;
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: all var(--transition-base);
  border: 1px solid var(--border-light);
}

.overview-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.card-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.card-content {
  flex: 1;
}

.card-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.card-label {
  color: var(--text-secondary);
  font-size: 14px;
  margin-top: 4px;
}

.card-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.trend {
  display: flex;
  align-items: center;
  gap: 2px;
  font-weight: 600;
}

.trend.up {
  color: var(--success-color);
}

.trend.down {
  color: var(--danger-color);
}

.trend-label {
  color: var(--text-tertiary);
}

/* 今日统计 */
.today-stats {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.stats-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.stats-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.stats-date {
  color: var(--text-tertiary);
  font-size: 13px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 12px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: white;
}

.stat-icon.new {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.closed {
  background: linear-gradient(135deg, #48bb78 0%, #38a169 100%);
}

.stat-icon.timeout {
  background: linear-gradient(135deg, #f56565 0%, #e53e3e 100%);
}

.stat-icon.time {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-value.danger {
  color: var(--danger-color);
}

.stat-label {
  font-size: 13px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

/* 内容网格 */
.content-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.content-card {
  background: var(--bg-primary);
  border-radius: 16px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
  overflow: hidden;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-light);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 深色主题卡片头部 */
[data-theme="dark"] .card-header {
  border-bottom-color: var(--border-color);
}

.card-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-body {
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

/* 工作负载 */
.workload-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.workload-item {
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
}

/* 深色主题工作负载项 */
[data-theme="dark"] .workload-item {
  background: var(--bg-tertiary);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.user-avatar {
  background: var(--primary-gradient);
  color: white;
  font-weight: 600;
}

.user-name {
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.user-stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.user-stats .stat {
  display: flex;
  align-items: center;
  gap: 4px;
}

.user-stats .stat.timeout {
  color: var(--danger-color);
}

.workload-bar {
  height: 6px;
  background: var(--border-light);
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: var(--primary-gradient);
  border-radius: 3px;
  transition: width var(--transition-base);
}

.bar-fill.warning {
  background: linear-gradient(135deg, #f6ad55 0%, #ed8936 100%);
}

.bar-fill.danger {
  background: linear-gradient(135deg, #fc8181 0%, #f56565 100%);
}

/* 超时列表 */
.timeout-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.timeout-item {
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.timeout-item:hover {
  background: var(--bg-hover);
}

/* 深色主题超时项 */
[data-theme="dark"] .timeout-item {
  background: var(--bg-tertiary);
}

[data-theme="dark"] .timeout-item:hover {
  background: var(--bg-hover);
}

.timeout-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.ticket-link {
  font-weight: 600;
  color: var(--primary-color);
  font-size: 15px;
}

.ticket-link:hover {
  text-decoration: underline;
}

.timeout-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.timeout-meta span {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 活动时间线 */
.activity-timeline {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-item {
  display: flex;
  gap: 16px;
}

.activity-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  margin-top: 6px;
  flex-shrink: 0;
}

.activity-dot.success {
  background: var(--success-color);
  box-shadow: 0 0 0 4px rgba(72, 187, 120, 0.2);
}

.activity-dot.primary {
  background: var(--primary-color);
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.2);
}

.activity-dot.warning {
  background: var(--warning-color);
  box-shadow: 0 0 0 4px rgba(237, 137, 54, 0.2);
}

.activity-dot.info {
  background: var(--info-color);
  box-shadow: 0 0 0 4px rgba(66, 153, 225, 0.2);
}

.activity-content {
  flex: 1;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-light);
}

/* 深色主题活动内容 */
[data-theme="dark"] .activity-content {
  border-bottom-color: var(--border-color);
}

.activity-item:last-child .activity-content {
  border-bottom: none;
  padding-bottom: 0;
}

.activity-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.activity-ticket {
  font-weight: 600;
  color: var(--text-primary);
}

.activity-desc {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
}

.activity-footer {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--text-tertiary);
}

.activity-footer span {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 响应式 */
@media (max-width: 768px) {
  .overview-cards {
    grid-template-columns: 1fr;
  }
  
  .content-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
