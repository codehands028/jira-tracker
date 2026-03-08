<template>
  <div class="tickets-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>工单列表</h1>
        <p>管理和跟踪所有工单</p>
      </div>
      <div class="header-right">
        <el-button
          v-if="canCreate"
          type="primary"
          @click="showCreateDialog = true"
        >
          <el-icon><Plus /></el-icon>
          创建工单
        </el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <el-select
          v-model="filterStatus"
          placeholder="状态筛选"
          clearable
          @change="fetchTickets"
          class="filter-item"
        >
          <el-option label="全部状态" value="" />
          <el-option label="处理中" value="processing">
            <el-tag type="primary" size="small">处理中</el-tag>
          </el-option>
          <el-option label="待复测" value="retesting">
            <el-tag type="warning" size="small">待复测</el-tag>
          </el-option>
          <el-option label="已关闭" value="closed">
            <el-tag type="success" size="small">已关闭</el-tag>
          </el-option>
        </el-select>

        <el-select
          v-model="filterTimeout"
          placeholder="超时筛选"
          clearable
          @change="fetchTickets"
          class="filter-item"
        >
          <el-option label="全部" value="" />
          <el-option label="超时工单" value="true">
            <el-tag type="danger" size="small">超时</el-tag>
          </el-option>
          <el-option label="未超时" value="false">
            <el-tag type="success" size="small">未超时</el-tag>
          </el-option>
        </el-select>

        <el-select
          v-model="filterPriority"
          placeholder="优先级筛选"
          clearable
          @change="fetchTickets"
          class="filter-item"
        >
          <el-option label="全部优先级" value="" />
          <el-option label="低" value="low" />
          <el-option label="中" value="medium" />
          <el-option label="高" value="high" />
          <el-option label="紧急" value="critical" />
        </el-select>

        <el-select
          v-if="isAdmin"
          v-model="filterUser"
          placeholder="处理人筛选"
          clearable
          @change="fetchTickets"
          class="filter-item"
        >
          <el-option label="全部处理人" value="" />
          <el-option
            v-for="user in users"
            :key="user.id"
            :label="user.name"
            :value="user.id"
          />
        </el-select>

        <el-input
          v-model="searchKeyword"
          placeholder="搜索工单编号或描述"
          clearable
          @keyup.enter="fetchTickets"
          class="search-input"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      
      <div class="filter-right">
        <el-button-group v-if="selectedTickets.length > 0" class="batch-button-group">
          <el-button type="primary" @click="showBatchDialog = true">
            批量操作 ({{ selectedTickets.length }})
          </el-button>
          <el-button type="primary" @click="handleSelectAll">
            {{ isAllSelected ? '取消全选' : '全选当前页' }}
          </el-button>
          <el-button type="primary" @click="handleClearSelection">
            清空选择
          </el-button>
        </el-button-group>
      </div>
    </div>

    <!-- 工单表格 -->
    <div class="table-container">
      <el-table
        ref="tableRef"
        :data="tickets"
        v-loading="loading"
        @selection-change="handleSelectionChange"
        class="modern-table"
        :row-class-name="getRowClassName"
        row-key="id"
      >
        <el-table-column type="selection" width="50" />
        
        <el-table-column label="工单编号" width="150">
          <template #default="{ row }">
            <router-link :to="`/tickets/${row.id}`" class="ticket-link">
              <el-icon class="link-icon"><Document /></el-icon>
              {{ row.jira_key }}
            </router-link>
          </template>
        </el-table-column>

        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="description">{{ row.description || '-' }}</span>
          </template>
        </el-table-column>

        <el-table-column label="优先级" width="100" align="center">
          <template #default="{ row }">
            <el-tag
              :type="getPriorityType(row.priority)"
              size="small"
              effect="dark"
            >
              {{ getPriorityLabel(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <div class="status-cell">
              <span class="status-dot" :class="row.status"></span>
              <span>{{ getStatusLabel(row.status) }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="当前处理人" width="120">
          <template #default="{ row }">
            <div class="handler-cell">
              <el-avatar :size="28" class="handler-avatar">
                {{ row.current_user?.name?.charAt(0) }}
              </el-avatar>
              <span>{{ row.current_user?.name }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="超时" width="80" align="center">
          <template #default="{ row }">
            <el-tag
              v-if="row.is_timeout"
              :type="row.timeout_level === 'severe' ? 'danger' : 'warning'"
              size="small"
              effect="dark"
            >
              {{ row.timeout_level === 'severe' ? '严重' : '超时' }}
            </el-tag>
            <span v-else class="no-timeout">-</span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="160" sortable>
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon><Clock /></el-icon>
              {{ formatTime(row.created_at) }}
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="viewTicket(row)"
            >
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchTickets"
          @current-change="fetchTickets"
        />
      </div>
    </div>

    <!-- 创建工单对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建工单"
      width="560px"
      class="modern-dialog"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="100px"
        label-position="top"
      >
        <el-form-item label="Jira工单编号" prop="jira_key">
          <el-input
            v-model="createForm.jira_key"
            placeholder="例如：PROJ-001"
          >
            <template #prefix>
              <el-icon><Document /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="Jira链接" prop="jira_url">
          <el-input
            v-model="createForm.jira_url"
            placeholder="https://jira.example.com/browse/PROJ-001"
          >
            <template #prefix>
              <el-icon><Link /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="问题描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="请简要描述问题"
          />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="优先级" prop="priority">
              <el-select v-model="createForm.priority" placeholder="选择优先级">
                <el-option label="低" value="low" />
                <el-option label="中" value="medium" />
                <el-option label="高" value="high" />
                <el-option label="紧急" value="critical" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="处理人" prop="current_user_id">
              <el-select
                v-model="createForm.current_user_id"
                placeholder="选择处理人"
                filterable
                @change="(val) => console.log('选择处理人:', val)"
              >
                <el-option
                  v-for="user in users"
                  :key="user.id"
                  :label="user.name"
                  :value="user.id"
                >
                  <div class="user-option">
                    <el-avatar :size="24" class="option-avatar">
                      {{ user.name.charAt(0) }}
                    </el-avatar>
                    <span>{{ user.name }}</span>
                    <el-tag size="small" :type="getRoleType(user.role)">
                      {{ getRoleLabel(user.role) }}
                    </el-tag>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleCreate"
          :loading="createLoading"
        >
          创建工单
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量操作对话框 -->
    <el-dialog
      v-model="showBatchDialog"
      title="批量操作"
      width="650px"
      class="modern-dialog"
    >
      <div class="batch-info">
        <el-icon class="info-icon"><InfoFilled /></el-icon>
        <span>已选择 <strong>{{ selectedTickets.length }}</strong> 个工单</span>
        <el-tag v-if="selectedTickets.length > 50" type="warning" size="small" style="margin-left: 8px;">
          建议单次操作不超过50条以确保性能
        </el-tag>
      </div>

      <!-- 工单概览 -->
      <div class="batch-overview">
        <div class="overview-header">
          <span>工单概览</span>
          <el-button link type="primary" size="small" @click="showAllTickets = !showAllTickets">
            {{ showAllTickets ? '收起' : '展开全部' }}
          </el-button>
        </div>
        <div class="overview-stats">
          <el-tag type="primary" size="small">处理中: {{ selectedStats.processing }}</el-tag>
          <el-tag type="warning" size="small">待复测: {{ selectedStats.retesting }}</el-tag>
          <el-tag type="danger" size="small">超时: {{ selectedStats.timeout }}</el-tag>
        </div>
        <div class="overview-tickets" v-if="showAllTickets">
          <el-tag
            v-for="ticket in selectedTickets"
            :key="ticket.id"
            size="small"
            :type="ticket.is_timeout ? 'danger' : 'info'"
            style="margin: 4px;"
          >
            {{ ticket.jira_key }}
          </el-tag>
        </div>
      </div>
      
      <div class="batch-actions">
        <el-button
          type="primary"
          @click="showBatchFlowDialog = true"
          class="batch-btn"
        >
          <el-icon><Right /></el-icon>
          批量流转
        </el-button>
        <el-button
          v-if="isAdmin"
          type="warning"
          @click="showBatchAssignDialog = true"
          class="batch-btn"
        >
          <el-icon><User /></el-icon>
          批量分配
        </el-button>
        <el-button
          v-if="canBatchClose"
          type="success"
          @click="showBatchCloseDialog = true"
          class="batch-btn"
        >
          <el-icon><CircleCheck /></el-icon>
          批量关闭
        </el-button>
      </div>
    </el-dialog>

    <!-- 批量流转对话框 -->
    <el-dialog
      v-model="showBatchFlowDialog"
      title="批量流转"
      width="500px"
      class="modern-dialog"
    >
      <el-form ref="batchFlowFormRef" :model="batchFlowForm" :rules="batchFlowRules" label-width="100px" label-position="top">
        <el-form-item label="流转说明" prop="content">
          <el-input
            v-model="batchFlowForm.content"
            type="textarea"
            :rows="3"
            placeholder="请输入流转说明"
          />
        </el-form-item>
        <el-form-item label="下一处理人" prop="to_user_id">
          <el-select
            v-model="batchFlowForm.to_user_id"
            placeholder="选择处理人"
            filterable
          >
            <el-option
              v-for="user in users"
              :key="user.id"
              :label="`${user.name} [${getRoleLabel(user.role)}]`"
              :value="user.id"
            >
              <span>{{ user.name }}</span>
              <el-tag :type="getRoleType(user.role)" size="small" style="margin-left: 8px;">
                {{ getRoleLabel(user.role) }}
              </el-tag>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchFlowDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleBatchFlow"
          :loading="batchLoading"
        >
          确认流转
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量分配对话框 -->
    <el-dialog
      v-model="showBatchAssignDialog"
      title="批量分配"
      width="500px"
      class="modern-dialog"
    >
      <el-form label-width="100px" label-position="top">
        <el-form-item label="分配给">
          <el-select
            v-model="batchAssignTo"
            placeholder="选择处理人"
            filterable
          >
            <el-option
              v-for="user in users"
              :key="user.id"
              :label="`${user.name} [${getRoleLabel(user.role)}]`"
              :value="user.id"
            >
              <span>{{ user.name }}</span>
              <el-tag :type="getRoleType(user.role)" size="small" style="margin-left: 8px;">
                {{ getRoleLabel(user.role) }}
              </el-tag>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="batchAssignContent"
            type="textarea"
            :rows="3"
            placeholder="请输入分配说明"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchAssignDialog = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleBatchAssign"
          :loading="batchLoading"
        >
          确认分配
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量关闭对话框 -->
    <el-dialog
      v-model="showBatchCloseDialog"
      title="批量关闭"
      width="500px"
      class="modern-dialog"
    >
      <el-form label-width="100px" label-position="top">
        <el-form-item label="关闭原因">
          <el-input
            v-model="batchCloseConclusion"
            type="textarea"
            :rows="3"
            placeholder="请输入关闭原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchCloseDialog = false">取消</el-button>
        <el-button
          type="success"
          @click="handleBatchClose"
          :loading="batchLoading"
        >
          确认关闭
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  getTickets,
  createTicket,
  batchAssignTickets,
  batchFlowTickets,
  batchCloseTickets,
  getUsers
} from '@/api'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const createLoading = ref(false)
const batchLoading = ref(false)
const tickets = ref([])
const users = ref([])
const selectedTickets = ref([])

const showCreateDialog = ref(false)
const showBatchDialog = ref(false)
const showBatchAssignDialog = ref(false)
const showBatchCloseDialog = ref(false)
const showBatchFlowDialog = ref(false)
const showAllTickets = ref(false)
const createFormRef = ref(null)
const batchFlowFormRef = ref(null)
const tableRef = ref(null)

const filterStatus = ref('')
const filterPriority = ref('')
const filterUser = ref('')
const filterTimeout = ref('')
const searchKeyword = ref('')

const batchAssignTo = ref(null)
const batchAssignContent = ref('')
const batchCloseConclusion = ref('')

const batchFlowForm = reactive({
  content: '',
  to_user_id: null
})

const batchFlowRules = {
  content: [{ required: true, message: '请输入流转说明', trigger: 'blur' }],
  to_user_id: [{ required: true, message: '请选择处理人', trigger: 'change' }]
}

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const createForm = reactive({
  jira_key: '',
  jira_url: '',
  description: '',
  priority: 'medium',
  current_user_id: undefined
})

const createRules = {
  jira_key: [
    { required: true, message: '请输入工单编号', trigger: 'blur' }
  ],
  jira_url: [
    { required: true, message: '请输入Jira链接', trigger: 'blur' }
  ],
  current_user_id: [
    { required: true, message: '请选择处理人', trigger: 'change' }
  ]
}

const canCreate = computed(() => {
  const role = userStore.role
  return role === 'test' || role === 'admin'
})

const isAdmin = computed(() => userStore.role === 'admin')

const canBatchClose = computed(() => {
  const role = userStore.role
  return role === 'test' || role === 'admin'
})

const isAllSelected = computed(() => {
  return tickets.value.length > 0 && selectedTickets.value.length === tickets.value.length
})

const selectedStats = computed(() => {
  const stats = {
    processing: 0,
    retesting: 0,
    timeout: 0
  }
  selectedTickets.value.forEach(ticket => {
    if (ticket.status === 'processing') stats.processing++
    if (ticket.status === 'retesting') stats.retesting++
    if (ticket.is_timeout) stats.timeout++
  })
  return stats
})

const fetchTickets = async () => {
  // 延迟显示 loading，避免快速请求时的闪烁
  let loadingTimer = null
  const showLoading = () => {
    loadingTimer = setTimeout(() => {
      loading.value = true
    }, 150)
  }
  showLoading()

  try {
    console.log('开始获取工单列表...')
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      status: filterStatus.value,
      priority: filterPriority.value
    }
    
    // 只有管理员才能按用户筛选
    if (isAdmin.value && filterUser.value) {
      params.current_user_id = filterUser.value
    }
    
    // 添加超时筛选参数
    if (filterTimeout.value) {
      params.is_timeout = filterTimeout.value
    }
    const data = await getTickets(params)
    console.log('获取工单列表成功:', data)
    tickets.value = data.list
    pagination.total = data.total
  } catch (error) {
    console.error('获取工单列表失败:', error)
    console.error('错误状态码:', error.response?.status)
    console.error('错误信息:', error.response?.data)
  } finally {
    if (loadingTimer) clearTimeout(loadingTimer)
    loading.value = false
  }
}

const fetchUsers = async () => {
  try {
    const data = await getUsers({ page: 1, page_size: 1000 })
    console.log('用户列表数据:', data.list)
    users.value = data.list
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

const handleSelectionChange = (selection) => {
  selectedTickets.value = selection
}

const handleSelectAll = () => {
  if (isAllSelected.value) {
    // 取消全选
    tableRef.value?.clearSelection()
  } else {
    // 全选当前页
    selectedTickets.value = [...tickets.value]
  }
}

const handleClearSelection = () => {
  tableRef.value?.clearSelection()
}

const handleCreate = async () => {
  try {
    await createFormRef.value.validate()
    createLoading.value = true
    
    await createTicket(createForm)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    
    Object.assign(createForm, {
      jira_key: '',
      jira_url: '',
      description: '',
      priority: 'medium',
      current_user_id: undefined
    })
    
    fetchTickets()
  } catch (error) {
    if (error !== false) {
      console.error('创建工单失败:', error)
    }
  } finally {
    createLoading.value = false
  }
}

const handleBatchAssign = async () => {
  if (!batchAssignTo.value) {
    ElMessage.warning('请选择处理人')
    return
  }

  if (selectedTickets.value.length > 100) {
    ElMessage.warning('单次批量操作建议不超过100条工单，以确保系统性能')
  }

  try {
    batchLoading.value = true
    const result = await batchAssignTickets({
      ticket_ids: selectedTickets.value.map(t => t.id),
      to_user_id: batchAssignTo.value,
      content: batchAssignContent.value
    })

    ElMessage.success(`成功分配 ${result.success_count} 个工单`)
    showBatchAssignDialog.value = false
    showBatchDialog.value = false
    fetchTickets()
  } catch (error) {
    console.error('批量分配失败:', error)
    // 错误提示已在 request.js 中统一处理
  } finally {
    batchLoading.value = false
  }
}

const handleBatchFlow = async () => {
  try {
    await batchFlowFormRef.value.validate()

    if (selectedTickets.value.length > 100) {
      ElMessage.warning('单次批量操作建议不超过100条工单，以确保系统性能')
    }

    batchLoading.value = true
    const result = await batchFlowTickets({
      ticket_ids: selectedTickets.value.map(t => t.id),
      to_user_id: batchFlowForm.to_user_id,
      content: batchFlowForm.content
    })

    ElMessage.success(`成功流转 ${result.success_count} 个工单`)
    showBatchFlowDialog.value = false
    showBatchDialog.value = false
    Object.assign(batchFlowForm, { content: '', to_user_id: null })
    fetchTickets()
  } catch (error) {
    if (error !== false) {
      console.error('批量流转失败:', error)
      // 错误提示已在 request.js 中统一处理
    }
  } finally {
    batchLoading.value = false
  }
}

const handleBatchClose = async () => {
  if (!batchCloseConclusion.value) {
    ElMessage.warning('请输入关闭原因')
    return
  }

  if (selectedTickets.value.length > 100) {
    ElMessage.warning('单次批量操作建议不超过100条工单，以确保系统性能')
  }

  try {
    batchLoading.value = true
    const result = await batchCloseTickets({
      ticket_ids: selectedTickets.value.map(t => t.id),
      conclusion: batchCloseConclusion.value
    })

    ElMessage.success(`成功关闭 ${result.success_count} 个工单`)
    showBatchCloseDialog.value = false
    showBatchDialog.value = false
    fetchTickets()
  } catch (error) {
    console.error('批量关闭失败:', error)
    // 错误提示已在 request.js 中统一处理
  } finally {
    batchLoading.value = false
  }
}

const viewTicket = (row) => {
  router.push(`/tickets/${row.id}`)
}

const getRowClassName = ({ row }) => {
  if (row.is_timeout) {
    return 'timeout-row'
  }
  return ''
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
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

onMounted(() => {
  fetchTickets()
  fetchUsers()
})
</script>

<style scoped>
.tickets-container {
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

/* 深色主题页面标题 */
[data-theme="dark"] .header-left h1 {
  color: var(--text-primary);
}

[data-theme="dark"] .header-left p {
  color: var(--text-tertiary);
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: var(--bg-primary);
  padding: 16px 20px;
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
}

.filter-left {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-item {
  width: 140px;
}

.search-input {
  width: 240px;
}

/* 表格容器 */
.table-container {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
}

.modern-table {
  border-radius: 12px;
  overflow: hidden;
}

/* 表格容器过渡动画 */
.modern-table :deep(.el-table__body-wrapper) {
  transition: opacity 0.2s ease;
}

/* 表格数据更新时淡入 */
.modern-table :deep(.el-table__row) {
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 工单链接 */
.ticket-link {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--primary-color);
  font-weight: 600;
  text-decoration: none;
}

.ticket-link:hover {
  text-decoration: underline;
}

.link-icon {
  font-size: 16px;
}

/* 描述 */
.description {
  color: var(--text-secondary);
}

/* 状态单元格 */
.status-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 100%;
}

.status-cell span {
  line-height: 1;
  display: flex;
  align-items: center;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.processing {
  background: var(--primary-color);
}

.status-dot.retesting {
  background: var(--warning-color);
}

.status-dot.closed {
  background: var(--success-color);
}

.status-dot.timeout {
  background: var(--danger-color);
}

/* 处理人单元格 */
.handler-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.handler-avatar {
  background: var(--primary-gradient);
  color: white;
  font-size: 12px;
}

/* 时间单元格 */
.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary);
  font-size: 13px;
}

.no-timeout {
  color: var(--text-disabled);
}

/* 超时行样式 */
:deep(.timeout-row) {
  background: rgba(245, 101, 101, 0.05) !important;
}

:deep(.timeout-row:hover > td) {
  background: rgba(245, 101, 101, 0.08) !important;
}

/* 深色模式超时行样式 */
[data-theme="dark"] :deep(.timeout-row) {
  background: rgba(239, 83, 80, 0.15) !important;
}

[data-theme="dark"] :deep(.timeout-row:hover > td) {
  background: rgba(239, 83, 80, 0.2) !important;
}

/* 分页 */
.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 用户选项 */
.user-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.option-avatar {
  background: var(--primary-gradient);
  color: white;
  font-size: 10px;
}

/* 批量操作 */
.batch-button-group {
  display: flex;
}

.batch-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(102, 126, 234, 0.05);
  border-radius: 8px;
  margin-bottom: 16px;
  color: var(--primary-color);
  font-weight: 500;
}

.batch-info strong {
  font-size: 16px;
  color: var(--primary-color);
}

.info-icon {
  font-size: 18px;
}

.batch-overview {
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 8px;
  margin-bottom: 16px;
}

.overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 500;
  color: var(--text-primary);
}

.overview-stats {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.overview-tickets {
  max-height: 150px;
  overflow-y: auto;
  padding: 8px;
  background: var(--bg-primary);
  border-radius: 6px;
}

.batch-actions {
  display: flex;
  gap: 12px;
}

.batch-btn {
  flex: 1;
  height: 60px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.batch-btn .el-icon {
  font-size: 20px;
}

/* 对话框 */
.modern-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-light);
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.modern-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid var(--border-light);
}

.modern-dialog :deep(.el-select) {
  width: 100%;
}

/* 响应式 */
@media (max-width: 768px) {
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-left {
    flex-direction: column;
  }

  .filter-item,
  .search-input {
    width: 100%;
  }
}
</style>
