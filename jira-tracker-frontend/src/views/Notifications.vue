<template>
  <div class="notifications-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>通知中心</h1>
        <p>查看所有通知和提醒</p>
      </div>
      <div class="header-right">
        <el-button @click="fetchNotifications" :icon="'Refresh'" circle />
      </div>
    </div>

    <!-- 通知列表 -->
    <div class="notifications-list">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        class="notification-card"
        :class="{ unread: !notification.is_read }"
      >
        <div class="notification-icon" :class="notification.type">
          <el-icon :size="24">
            <component :is="getTypeIcon(notification.type)" />
          </el-icon>
        </div>
        <div class="notification-content">
          <div class="notification-header">
            <h4>{{ notification.title }}</h4>
            <div class="notification-tags">
              <el-tag :type="getNotificationType(notification.type)" size="small">
                {{ getNotificationTypeLabel(notification.type) }}
              </el-tag>
              <el-tag
                v-if="!notification.is_read"
                type="danger"
                size="small"
                effect="dark"
              >
                未读
              </el-tag>
            </div>
          </div>
          <p class="notification-text">{{ notification.content }}</p>
          <div class="notification-footer">
            <span class="time">
              <el-icon><Clock /></el-icon>
              {{ formatTime(notification.created_at) }}
            </span>
            <el-button
              v-if="!notification.is_read"
              link
              type="primary"
              @click="handleMarkAsRead(notification)"
            >
              标记已读
            </el-button>
            <router-link
              v-if="notification.ticket_id"
              :to="`/tickets/${notification.ticket_id}`"
              class="view-ticket"
            >
              查看工单
            </router-link>
          </div>
        </div>
      </div>
      
      <el-empty v-if="notifications.length === 0 && !loading" description="暂无通知" />
    </div>

    <!-- 分页 -->
    <div class="pagination-container" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchNotifications"
        @current-change="fetchNotifications"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getNotifications, markNotificationAsRead } from '@/api'
import { useNotificationStore } from '@/stores/notification'
import dayjs from 'dayjs'

const loading = ref(false)
const notifications = ref([])
const notificationStore = useNotificationStore()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const fetchNotifications = async () => {
  try {
    loading.value = true
    const data = await getNotifications({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    notifications.value = data.list
    pagination.total = data.total
  } catch (error) {
    console.error('获取通知列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleMarkAsRead = async (notification) => {
  try {
    await markNotificationAsRead(notification.id)
    ElMessage.success('标记成功')
    // 更新notification store中的未读计数
    notificationStore.fetchNotifications()
    fetchNotifications()
  } catch (error) {
    console.error('标记通知失败:', error)
  }
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

const getTypeIcon = (type) => {
  const map = {
    timeout: 'Warning',
    flow: 'Right',
    close: 'CircleCheck'
  }
  return map[type] || 'Bell'
}

const getNotificationType = (type) => {
  const map = {
    timeout: 'danger',
    flow: 'primary',
    close: 'success'
  }
  return map[type] || 'info'
}

const getNotificationTypeLabel = (type) => {
  const map = {
    timeout: '超时提醒',
    flow: '工单流转',
    close: '工单关闭'
  }
  return map[type] || '通知'
}

onMounted(() => {
  fetchNotifications()
})
</script>

<style scoped>
.notifications-container {
  padding: 0;
}

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

.notifications-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.notification-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
  display: flex;
  gap: 16px;
  transition: all var(--transition-base);
}

.notification-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.notification-card.unread {
  border-left: 4px solid var(--primary-color);
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.02) 0%, rgba(118, 75, 162, 0.02) 100%);
}

.notification-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.notification-icon.timeout {
  background: linear-gradient(135deg, rgba(245, 101, 101, 0.1) 0%, rgba(229, 62, 62, 0.1) 100%);
  color: var(--danger-color);
}

.notification-icon.flow {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  color: var(--primary-color);
}

.notification-icon.close {
  background: linear-gradient(135deg, rgba(72, 187, 120, 0.1) 0%, rgba(56, 161, 105, 0.1) 100%);
  color: var(--success-color);
}

.notification-content {
  flex: 1;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.notification-tags {
  display: flex;
  gap: 8px;
}

.notification-header h4 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.notification-text {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 12px;
  line-height: 1.5;
}

.notification-footer {
  display: flex;
  align-items: center;
  gap: 16px;
}

.time {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary);
  font-size: 13px;
}

.view-ticket {
  color: var(--primary-color);
  font-size: 13px;
}

.view-ticket:hover {
  text-decoration: underline;
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  background: var(--bg-primary);
  padding: 16px 20px;
  border-radius: 12px;
}
</style>
