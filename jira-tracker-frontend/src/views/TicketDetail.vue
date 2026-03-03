<template>
  <div class="ticket-detail-container">
    <!-- 返回按钮 -->
    <div class="back-bar">
      <el-button @click="handleBack" link>
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </el-button>
    </div>

    <div class="detail-content" v-if="ticket">
      <!-- 工单信息卡片 -->
      <div class="info-card">
        <div class="card-header">
          <div class="header-left">
            <h2>{{ ticket.jira_key }}</h2>
            <el-tag :type="getStatusType(ticket.status)" effect="dark">
              {{ getStatusLabel(ticket.status) }}
            </el-tag>
            <el-tag v-if="ticket.is_timeout" type="danger" effect="dark">
              {{ ticket.timeout_level === 'severe' ? '严重超时' : '已超时' }}
            </el-tag>
          </div>
          <el-link :href="ticket.jira_url" target="_blank" type="primary">
            <el-icon><Link /></el-icon>
            在Jira中查看
          </el-link>
        </div>

        <div class="info-grid">
          <div class="info-item">
            <span class="label">优先级</span>
            <el-tag :type="getPriorityType(ticket.priority)" effect="dark">
              {{ getPriorityLabel(ticket.priority) }}
            </el-tag>
          </div>
          
          <div class="info-item">
            <span class="label">当前处理人</span>
            <div class="handler-info">
              <el-avatar :size="28" class="handler-avatar">
                {{ ticket.current_user?.name?.charAt(0) }}
              </el-avatar>
              <span>{{ ticket.current_user?.name }}</span>
            </div>
          </div>
          
          <div class="info-item full">
            <span class="label">问题描述</span>
            <p class="description">{{ ticket.description || '暂无描述' }}</p>
          </div>
          
          <div class="info-item">
            <span class="label">创建时间</span>
            <span class="value">{{ formatTime(ticket.created_at) }}</span>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="action-buttons">
          <el-button
            v-if="canFlow"
            type="primary"
            @click="showFlowDialog = true"
          >
            <el-icon><Right /></el-icon>
            流转工单
          </el-button>
          <el-button
            v-if="isTestOrAdmin && (ticket.status === 'retesting' || ticket.status === 'timeout')"
            type="success"
            @click="showRetestDialog = true"
          >
            <el-icon><CircleCheck /></el-icon>
            复测工单
          </el-button>
          <el-button
            v-if="isTestOrAdmin && (ticket.status === 'retesting' || ticket.status === 'processing' || ticket.status === 'timeout')"
            type="danger"
            @click="showCloseDialog = true"
          >
            <el-icon><Close /></el-icon>
            关闭工单
          </el-button>
        </div>
      </div>

      <!-- 流转历史 -->
      <div class="flow-card">
        <div class="card-header">
          <h3>流转历史</h3>
          <el-tag type="info" size="small">{{ flows.length }} 条记录</el-tag>
        </div>

        <div class="flow-timeline">
          <div
            v-for="flow in flows"
            :key="flow.id"
            class="flow-item"
            :class="{ timeout: flow.is_timeout }"
          >
            <div class="flow-dot" :class="{ timeout: flow.is_timeout }"></div>
            <div class="flow-content">
              <div class="flow-header">
                <div class="users">
                  <el-avatar :size="32" class="flow-avatar">
                    {{ flow.from_user?.name?.charAt(0) }}
                  </el-avatar>
                  <el-icon class="arrow"><Right /></el-icon>
                  <el-avatar :size="32" class="flow-avatar to">
                    {{ flow.to_user?.name?.charAt(0) }}
                  </el-avatar>
                </div>
                <el-tag v-if="flow.is_timeout" type="danger" size="small">
                  {{ flow.timeout_level === 'severe' ? '严重超时' : '超时' }}
                </el-tag>
              </div>
              <p class="flow-text">{{ flow.content }}</p>
              <div class="flow-meta">
                <span v-if="flow.process_time" class="process-time">
                  <el-icon><Clock /></el-icon>
                  处理时长: {{ formatDuration(flow.process_time) }}
                </span>
                <span class="time">{{ formatTime(flow.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 流转对话框 -->
    <el-dialog v-model="showFlowDialog" title="流转工单" width="500px" class="modern-dialog">
      <el-form ref="flowFormRef" :model="flowForm" :rules="flowRules" label-width="100px">
        <el-form-item label="处理说明" prop="content">
          <el-input v-model="flowForm.content" type="textarea" :rows="4" placeholder="请输入处理说明" />
        </el-form-item>
        <el-form-item label="下一处理人" prop="to_user_id">
          <el-select v-model="flowForm.to_user_id" placeholder="请选择处理人" filterable>
            <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showFlowDialog = false">取消</el-button>
        <el-button type="primary" @click="handleFlow" :loading="flowLoading">确认流转</el-button>
      </template>
    </el-dialog>

    <!-- 复测对话框 -->
    <el-dialog v-model="showRetestDialog" title="复测工单" width="500px" class="modern-dialog">
      <el-form ref="retestFormRef" :model="retestForm" :rules="retestRules" label-width="100px">
        <el-form-item label="复测结果" prop="passed">
          <el-radio-group v-model="retestForm.passed">
            <el-radio :label="true">通过</el-radio>
            <el-radio :label="false">不通过</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="复测结论" prop="content">
          <el-input v-model="retestForm.content" type="textarea" :rows="4" placeholder="请输入复测结论" />
        </el-form-item>
        <el-form-item v-if="!retestForm.passed" label="处理人" prop="to_user_id">
          <el-select v-model="retestForm.to_user_id" placeholder="请选择处理人" filterable>
            <el-option v-for="user in users" :key="user.id" :label="user.name" :value="user.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRetestDialog = false">取消</el-button>
        <el-button type="primary" @click="handleRetest" :loading="retestLoading">确认</el-button>
      </template>
    </el-dialog>

    <!-- 关闭对话框 -->
    <el-dialog v-model="showCloseDialog" title="关闭工单" width="500px" class="modern-dialog">
      <el-form ref="closeFormRef" :model="closeForm" :rules="closeRules" label-width="100px">
        <el-form-item label="关闭结论" prop="conclusion">
          <el-input v-model="closeForm.conclusion" type="textarea" :rows="4" placeholder="请输入关闭结论" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCloseDialog = false">取消</el-button>
        <el-button type="danger" @click="handleClose" :loading="closeLoading">确认关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getTicketDetail, flowTicket, retestTicket, closeTicket, getUsers } from '@/api'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'

dayjs.extend(duration)

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const flowLoading = ref(false)
const retestLoading = ref(false)
const closeLoading = ref(false)
const showFlowDialog = ref(false)
const showRetestDialog = ref(false)
const showCloseDialog = ref(false)
const flowFormRef = ref(null)
const retestFormRef = ref(null)
const closeFormRef = ref(null)

const ticket = ref(null)
const flows = ref([])
const users = ref([])

const isTestOrAdmin = computed(() => {
  const role = userStore.role
  return role === 'test' || role === 'admin'
})

const canFlow = computed(() => {
  return ticket.value && 
         ticket.value.current_user_id === userStore.userInfo?.id && 
         ticket.value.status !== 'closed'
})

const flowForm = reactive({
  content: '',
  to_user_id: null
})

const flowRules = {
  content: [{ required: true, message: '请输入处理说明', trigger: 'blur' }],
  to_user_id: [{ required: true, message: '请选择处理人', trigger: 'change' }]
}

const retestForm = reactive({
  passed: true,
  content: '',
  to_user_id: null
})

const retestRules = {
  content: [{ required: true, message: '请输入复测结论', trigger: 'blur' }],
  to_user_id: [
    {
      validator: (rule, value, callback) => {
        if (!retestForm.passed && !value) {
          callback(new Error('请选择处理人'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

const closeForm = reactive({
  conclusion: ''
})

const closeRules = {
  conclusion: [{ required: true, message: '请输入关闭结论', trigger: 'blur' }]
}

const fetchTicketDetail = async () => {
  try {
    loading.value = true
    const data = await getTicketDetail(route.params.id)
    ticket.value = data.ticket
    flows.value = data.flows
  } catch (error) {
    console.error('获取工单详情失败:', error)
  } finally {
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const data = await getUsers({ page: 1, page_size: 1000 })
    users.value = data.list
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

const handleBack = () => {
  router.back()
}

const handleFlow = async () => {
  try {
    await flowFormRef.value.validate()
    flowLoading.value = true
    await flowTicket(route.params.id, flowForm)
    ElMessage.success('流转成功')
    showFlowDialog.value = false
    Object.assign(flowForm, { content: '', to_user_id: null })
    fetchTicketDetail()
  } catch (error) {
    if (error !== false) {
      console.error('流转工单失败:', error)
    }
  } finally {
    flowLoading.value = false
  }
}

const handleRetest = async () => {
  try {
    await retestFormRef.value.validate()
    retestLoading.value = true
    await retestTicket(route.params.id, retestForm)
    ElMessage.success('操作成功')
    showRetestDialog.value = false
    Object.assign(retestForm, { passed: true, content: '', to_user_id: null })
    fetchTicketDetail()
  } catch (error) {
    if (error !== false) {
      console.error('复测工单失败:', error)
    }
  } finally {
    retestLoading.value = false
  }
}

const handleClose = async () => {
  try {
    await closeFormRef.value.validate()
    closeLoading.value = true
    await closeTicket(route.params.id, closeForm)
    ElMessage.success('关闭成功')
    showCloseDialog.value = false
    closeForm.conclusion = ''
    fetchTicketDetail()
  } catch (error) {
    if (error !== false) {
      console.error('关闭工单失败:', error)
    }
  } finally {
    closeLoading.value = false
  }
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const dur = dayjs.duration(seconds, 'seconds')
  const days = dur.days()
  const hours = dur.hours()
  const minutes = dur.minutes()

  let result = ''
  if (days > 0) result += `${days}天`
  if (hours > 0) result += `${hours}小时`
  if (minutes > 0) result += `${minutes}分钟`
  if (!result) result = '不到1分钟'

  return result
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
    timeout: '超时'
  }
  return map[status] || status
}

const getPriorityType = (priority) => {
  const map = {
    low: 'info',
    medium: 'success',
    high: 'warning',
    critical: 'danger'
  }
  return map[priority] || ''
}

const getPriorityLabel = (priority) => {
  const map = {
    low: '低',
    medium: '中',
    high: '高',
    critical: '紧急'
  }
  return map[priority] || priority
}

onMounted(() => {
  fetchTicketDetail()
  fetchUsers()
})
</script>

<style scoped>
.ticket-detail-container {
  padding: 0;
  max-width: 1200px;
  margin: 0 auto;
}

.back-bar {
  margin-bottom: 20px;
}

.detail-content {
  display: grid;
  gap: 20px;
}

/* 信息卡片 */
.info-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 24px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.info-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-light);
}

/* 深色主题信息卡片 */
[data-theme="dark"] .info-card {
  background: var(--bg-primary);
  border-color: var(--border-color);
}

[data-theme="dark"] .info-card .card-header {
  border-bottom-color: var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-left h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item.full {
  grid-column: span 2;
}

.info-item .label {
  font-size: 13px;
  color: var(--text-tertiary);
}

.info-item .value {
  color: var(--text-primary);
  font-weight: 500;
}

.description {
  color: var(--text-secondary);
  line-height: 1.6;
}

.handler-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.handler-avatar {
  background: var(--primary-gradient);
  color: white;
  font-size: 12px;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 12px;
  padding-top: 20px;
  border-top: 1px solid var(--border-light);
}

/* 流转卡片 */
.flow-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 24px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.flow-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.flow-card h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.flow-timeline {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.flow-item {
  display: flex;
  gap: 16px;
  position: relative;
  padding-left: 20px;
}

.flow-item::before {
  content: '';
  position: absolute;
  left: 6px;
  top: 32px;
  bottom: -20px;
  width: 2px;
  background: var(--border-light);
}

/* 深色主题流转卡片 */
[data-theme="dark"] .flow-card {
  background: var(--bg-primary);
  border-color: var(--border-color);
}

[data-theme="dark"] .flow-item::before {
  background: var(--border-color);
}

.flow-item:last-child::before {
  display: none;
}

.flow-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--primary-color);
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.flow-dot.timeout {
  background: var(--danger-color);
}

.flow-content {
  flex: 1;
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 12px;
}

.flow-item.timeout .flow-content {
  background: rgba(245, 101, 101, 0.05);
}

/* 深色主题流转内容 */
[data-theme="dark"] .flow-content {
  background: var(--bg-tertiary);
}

[data-theme="dark"] .flow-item.timeout .flow-content {
  background: rgba(248, 113, 113, 0.15);
}

.flow-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.users {
  display: flex;
  align-items: center;
  gap: 8px;
}

.flow-avatar {
  background: var(--primary-gradient);
  color: white;
  font-size: 12px;
}

.flow-avatar.to {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.arrow {
  color: var(--text-tertiary);
}

.flow-text {
  color: var(--text-secondary);
  margin-bottom: 12px;
  line-height: 1.5;
}

.flow-meta {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--text-tertiary);
}

.process-time {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* 对话框 */
.modern-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-light);
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.modern-dialog :deep(.el-select) {
  width: 100%;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .info-item.full {
    grid-column: span 1;
  }
}
</style>
