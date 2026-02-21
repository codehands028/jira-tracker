<template>
  <div class="statistics-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>数据统计</h1>
        <p>查看工单处理效率分析</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-tabs v-model="activeTab" class="modern-tabs">
      <!-- 人员维度统计 -->
      <el-tab-pane label="人员维度" name="user">
        <div class="stats-grid">
          <div
            v-for="user in statistics.user_stats"
            :key="user.id"
            class="user-card"
          >
            <div class="user-header">
              <el-avatar :size="48" class="user-avatar">
                {{ user.name?.charAt(0) }}
              </el-avatar>
              <div class="user-info">
                <h4>{{ user.name }}</h4>
                <el-tag :type="getRoleType(user.role)" size="small">
                  {{ getRoleLabel(user.role) }}
                </el-tag>
              </div>
            </div>
            
            <div class="stats-items">
              <div class="stat-item">
                <span class="stat-value">{{ user.ticket_num }}</span>
                <span class="stat-label">待处理</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ user.total_tickets }}</span>
                <span class="stat-label">已处理</span>
              </div>
              <div class="stat-item">
                <span class="stat-value">{{ formatDuration(user.avg_process_time) }}</span>
                <span class="stat-label">平均时长</span>
              </div>
              <div class="stat-item">
                <span class="stat-value danger">{{ user.timeout_count }}</span>
                <span class="stat-label">超时次数</span>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- 工单维度统计 -->
      <el-tab-pane label="工单维度" name="ticket">
        <div class="table-container ticket-table">
          <el-table :data="statistics.ticket_stats" stripe class="modern-table" style="width: 100%">
            <el-table-column label="工单编号" prop="jira_key" min-width="150">
              <template #default="{ row }">
                <router-link :to="`/tickets/${row.id}`" class="ticket-link">
                  {{ row.jira_key }}
                </router-link>
              </template>
            </el-table-column>
            
            <el-table-column label="状态" prop="status" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small" effect="dark">
                  {{ getStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="总处理时长" prop="total_duration" width="150" align="center">
              <template #default="{ row }">
                {{ formatDuration(row.total_duration) }}
              </template>
            </el-table-column>
            
            <el-table-column label="流转次数" width="100" align="center">
              <template #default="{ row }">
                <el-tag type="info" size="small">{{ row.flow_count }}次</el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="创建时间" prop="created_at" width="180" align="center">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getStatistics } from '@/api'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'

dayjs.extend(duration)

const activeTab = ref('user')
const statistics = reactive({
  user_stats: [],
  ticket_stats: []
})

const fetchStatistics = async () => {
  try {
    const data = await getStatistics()
    statistics.user_stats = data.user_stats
    statistics.ticket_stats = data.ticket_stats
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
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
  if (minutes > 0) result += `${minutes}分钟`
  if (!result) result = '不到1分钟'

  return result
}

const getRoleType = (role) => {
  const map = {
    admin: 'danger',
    test: 'warning',
    dev: 'primary'
  }
  return map[role] || ''
}

const getRoleLabel = (role) => {
  const map = {
    admin: '管理员',
    test: '测试',
    dev: '研发'
  }
  return map[role] || role
}

const getStatusType = (status) => {
  const map = {
    processing: 'primary',
    retesting: 'warning',
    closed: 'success',
    timeout: 'danger'
  }
  return map[status] || ''
}

const getStatusLabel = (status) => {
  const map = {
    processing: '处理中',
    retesting: '待复测',
    closed: '已关闭',
    timeout: '已超时'
  }
  return map[status] || status
}

onMounted(() => {
  fetchStatistics()
})
</script>

<style scoped>
.statistics-container {
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

.modern-tabs {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.user-card {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 20px;
  transition: all var(--transition-base);
}

.user-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.user-avatar {
  background: var(--primary-gradient);
  color: white;
  font-weight: 600;
}

.user-info h4 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.stats-items {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.stat-item {
  text-align: center;
  padding: 12px;
  background: white;
  border-radius: 8px;
}

.stat-value {
  display: block;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.stat-value.danger {
  color: var(--danger-color);
}

.stat-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.table-container {
  border-radius: 12px;
  overflow: hidden;
}

.ticket-table {
  width: 100%;
  background: var(--bg-secondary);
  padding: 16px;
  border-radius: 12px;
}

.ticket-table .el-table {
  width: 100% !important;
}

.ticket-link {
  color: var(--primary-color);
  font-weight: 600;
}

.ticket-link:hover {
  text-decoration: underline;
}
</style>
