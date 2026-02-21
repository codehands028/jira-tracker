<template>
  <div class="timeout-rules-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>超时规则</h1>
        <p>配置工单超时提醒规则</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新增规则
        </el-button>
      </div>
    </div>

    <!-- 当前生效规则提示 -->
    <div v-if="activeRule" class="active-rule-banner">
      <div class="banner-icon">
        <el-icon :size="24"><CircleCheck /></el-icon>
      </div>
      <div class="banner-content">
        <h4>当前生效规则：{{ activeRule.name }}</h4>
        <p>
          普通超时：<strong>{{ activeRule.normal_limit_hours }}小时</strong>
          <span class="divider">|</span>
          严重超时：<strong>{{ activeRule.severe_limit_hours }}小时</strong>
        </p>
      </div>
    </div>

    <!-- 规则列表 -->
    <div class="rules-grid">
      <div
        v-for="rule in rules"
        :key="rule.id"
        class="rule-card"
        :class="{ active: rule.is_active }"
      >
        <div class="rule-header">
          <h3>{{ rule.name }}</h3>
          <el-tag
            :type="rule.is_active ? 'success' : 'info'"
            effect="dark"
            size="small"
          >
            {{ rule.is_active ? '生效中' : '未生效' }}
          </el-tag>
        </div>
        
        <div class="rule-times">
          <div class="time-item">
            <span class="time-label">普通超时</span>
            <el-tooltip :content="`${formatHours(rule.normal_limit)}小时`" placement="top">
              <span class="time-value">{{ formatHours(rule.normal_limit) }}小时</span>
            </el-tooltip>
          </div>
          <div class="time-item severe">
            <span class="time-label">严重超时</span>
            <el-tooltip :content="`${formatHours(rule.severe_limit)}小时`" placement="top">
              <span class="time-value">{{ formatHours(rule.severe_limit) }}小时</span>
            </el-tooltip>
          </div>
        </div>
        
        <div class="rule-footer">
          <span class="create-time">
            <el-icon><Clock /></el-icon>
            {{ formatTime(rule.created_at) }}
          </span>
          <div class="rule-actions">
            <el-button
              v-if="!rule.is_active"
              type="success"
              size="small"
              text
              @click="handleSetActive(rule)"
            >
              设为生效
            </el-button>
            
            <el-button
              type="success"
              size="small"
              text
              class="edit-btn"
              @click="handleEdit(rule)"
            >
              编辑
            </el-button>
            <el-button
              v-if="!rule.is_active"
              type="danger"
              size="small"
              text
              @click="handleDelete(rule)"
            >
              删除
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建规则对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="新增超时规则"
      width="480px"
      class="modern-dialog"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="createForm.name" placeholder="例如：紧急规则" />
        </el-form-item>
        
        <el-form-item label="普通超时" prop="normal_limit_hours">
          <el-input-number
            v-model="createForm.normal_limit_hours"
            :min="1"
            :max="168"
          />
          <span class="unit">小时</span>
        </el-form-item>
        
        <el-form-item label="严重超时" prop="severe_limit_hours">
          <el-input-number
            v-model="createForm.severe_limit_hours"
            :min="1"
            :max="168"
          />
          <span class="unit">小时</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="createLoading">
          创建规则
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑规则对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑超时规则"
      width="480px"
      class="modern-dialog"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        
        <el-form-item label="普通超时" prop="normal_limit_hours">
          <el-input-number
            v-model="editForm.normal_limit_hours"
            :min="1"
            :max="168"
          />
          <span class="unit">小时</span>
        </el-form-item>
        
        <el-form-item label="严重超时" prop="severe_limit_hours">
          <el-input-number
            v-model="editForm.severe_limit_hours"
            :min="1"
            :max="168"
          />
          <span class="unit">小时</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdate" :loading="editLoading">
          更新规则
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import {
  getTimeoutRules,
  getActiveTimeoutRule,
  createTimeoutRule,
  updateTimeoutRule,
  setActiveTimeoutRule,
  setInactiveTimeoutRule,
  deleteTimeoutRule
} from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'

const rules = ref([])
const activeRule = ref(null)
const createLoading = ref(false)
const editLoading = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const createFormRef = ref(null)
const editFormRef = ref(null)

const createForm = reactive({
  name: '',
  normal_limit_hours: 24,
  severe_limit_hours: 48
})

const editForm = reactive({
  id: null,
  name: '',
  normal_limit_hours: 24,
  severe_limit_hours: 48
})

const formRules = {
  name: [
    { required: true, message: '请输入规则名称', trigger: 'blur' }
  ],
  normal_limit_hours: [
    { required: true, message: '请输入普通超时时间', trigger: 'blur' }
  ],
  severe_limit_hours: [
    { required: true, message: '请输入严重超时时间', trigger: 'blur' }
  ]
}

const fetchRules = async () => {
  try {
    const data = await getTimeoutRules()
    // 使用Map去重，保留ID最大的记录
    const ruleMap = new Map()
    ;(data.list || []).forEach(rule => {
      const id = Number(rule.id)
      if (!ruleMap.has(id) || ruleMap.get(id).id < id) {
        ruleMap.set(id, {
          ...rule,
          id: id
        })
      }
    })
    rules.value = Array.from(ruleMap.values())
  } catch (error) {
    console.error('获取超时规则失败:', error)
  }
}

const fetchActiveRule = async () => {
  try {
    const data = await getActiveTimeoutRule()
    activeRule.value = {
      ...data,
      normal_limit_hours: formatHours(data.normal_limit),
      severe_limit_hours: formatHours(data.severe_limit)
    }
  } catch (error) {
    console.error('获取生效规则失败:', error)
  }
}

const handleCreate = async () => {
  try {
    await createFormRef.value.validate()
    
    if (createForm.normal_limit_hours >= createForm.severe_limit_hours) {
      ElMessage.warning('严重超时时间必须大于普通超时时间')
      return
    }

    createLoading.value = true
    await createTimeoutRule(createForm)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    
    Object.assign(createForm, {
      name: '',
      normal_limit_hours: 24,
      severe_limit_hours: 48
    })
    
    fetchRules()
  } catch (error) {
    if (error !== false) {
      console.error('创建超时规则失败:', error)
    }
  } finally {
    createLoading.value = false
  }
}

const handleEdit = (rule) => {
  Object.assign(editForm, {
    id: Number(rule.id),
    name: rule.name,
    normal_limit_hours: formatHours(rule.normal_limit),
    severe_limit_hours: formatHours(rule.severe_limit)
  })
  showEditDialog.value = true
}

const handleUpdate = async () => {
  try {
    await editFormRef.value.validate()
    
    if (editForm.normal_limit_hours >= editForm.severe_limit_hours) {
      ElMessage.warning('严重超时时间必须大于普通超时时间')
      return
    }

    editLoading.value = true
    await updateTimeoutRule(editForm.id, {
      name: editForm.name,
      normal_limit_hours: editForm.normal_limit_hours,
      severe_limit_hours: editForm.severe_limit_hours
    })
    
    ElMessage.success('更新成功')
    showEditDialog.value = false
    fetchRules()
    fetchActiveRule()
  } catch (error) {
    if (error !== false) {
      console.error('更新超时规则失败:', error)
    }
  } finally {
    editLoading.value = false
  }
}

const handleSetActive = async (rule) => {
  try {
    await ElMessageBox.confirm(`确定要将规则 "${rule.name}" 设为生效吗？`, '提示', {
      type: 'warning'
    })

    await setActiveTimeoutRule(rule.id)
    ElMessage.success('设置成功')
    fetchRules()
    fetchActiveRule()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('设置生效规则失败:', error)
    }
  }
}

const handleSetInactive = async (rule) => {
  try {
    await ElMessageBox.confirm(`确定要禁用规则 "${rule.name}" 吗？禁用后将不再生效。`, '提示', {
      type: 'warning'
    })

    await setInactiveTimeoutRule(rule.id)
    ElMessage.success('禁用成功')
    fetchRules()
    fetchActiveRule()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('禁用超时规则失败:', error)
    }
  }
}

const handleDelete = async (rule) => {
  try {
    await ElMessageBox.confirm(`确定要删除规则 "${rule.name}" 吗？`, '提示', {
      type: 'warning'
    })

    await deleteTimeoutRule(rule.id)
    ElMessage.success('删除成功')
    fetchRules()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除超时规则失败:', error)
    }
  }
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD')
}

const formatHours = (duration) => {
  if (!duration) return 0
  // time.Duration在JSON中通常以纳秒为单位返回
  // 如果值很大（超过1000000000），说明是纳秒，需要转换为小时
  if (duration > 1000000000) {
    return Math.round(duration / 3600000000000)
  }
  // 否则假设是秒，转换为小时
  return Math.round(duration / 3600)
}

onMounted(() => {
  fetchRules()
  fetchActiveRule()
})
</script>

<style scoped>
.timeout-rules-container {
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

.active-rule-banner {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 1px solid rgba(102, 126, 234, 0.2);
  border-radius: 12px;
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.banner-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: var(--primary-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.banner-content h4 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.banner-content p {
  color: var(--text-secondary);
  font-size: 14px;
}

.banner-content strong {
  color: var(--primary-color);
}

.divider {
  margin: 0 12px;
  color: var(--text-disabled);
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.rule-card {
  background: white;
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow-card);
  border: 2px solid var(--border-light);
  transition: all var(--transition-base);
  box-sizing: border-box;
  overflow: hidden;
}

.rule-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.rule-card.active {
  border-color: var(--primary-color);
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.02) 0%, rgba(118, 75, 162, 0.02) 100%);
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.rule-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.rule-times {
  display: flex;
  gap: 16px;
  margin-bottom: 32px;
}

.time-item {
  flex: 1;
  padding: 16px 12px;
  background: var(--bg-secondary);
  border-radius: 12px;
  text-align: center;
  min-width: 0;
  overflow: hidden;
}

.time-item.severe {
  background: rgba(245, 101, 101, 0.05);
}

.time-label {
  display: block;
  font-size: 12px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}

.time-value {
  display: block;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  word-wrap: break-word;
  overflow-wrap: break-word;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time-item.severe .time-value {
  color: var(--danger-color);
}

.rule-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16px;
  gap: 8px;
  border-top: 1px solid var(--border-light);
  width: 100%;
  box-sizing: border-box;
  flex-wrap: nowrap;
}

.create-time {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary);
  font-size: 12px;
  flex-shrink: 0;
}

.rule-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.rule-actions .el-button {
  padding: 4px 8px;
  font-size: 12px;
}

.rule-actions .el-button--text {
  background: transparent !important;
}

.rule-actions .el-button--primary {
  background: transparent !important;
  color: var(--el-color-primary) !important;
}

.rule-actions .el-button--success {
  background: transparent !important;
}

.rule-actions .el-button--success.el-button--text {
  color: #409EFF !important;
}

.rule-actions .el-button--success.el-button--text:hover {
  color: #66b1ff !important;
}

.rule-actions .el-button--success.el-button--text:active {
  color: #3a8ee6 !important;
}

.rule-actions .el-button.edit-btn {
  color: #409EFF !important;
}

.rule-actions .el-button.edit-btn:hover {
  color: #66b1ff !important;
}

.rule-actions .el-button.edit-btn:active {
  color: #3a8ee6 !important;
}

.modern-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-light);
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.unit {
  margin-left: 8px;
  color: var(--text-tertiary);
}
</style>
