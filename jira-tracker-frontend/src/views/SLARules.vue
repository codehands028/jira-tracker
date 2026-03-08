<template>
  <div class="sla-rules-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>时限规则管理</h1>
        <p>配置工单处理时限规则</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新建规则
        </el-button>
      </div>
    </div>

    <!-- 规则列表 -->
    <div class="table-container">
      <el-table :data="rules" v-loading="loading" class="modern-table">
        <el-table-column prop="name" label="规则名称" min-width="150" />
        
        <el-table-column label="适用范围" min-width="200">
          <template #default="{ row }">
            <div class="scope-tags">
              <el-tag v-if="row.priority" :type="getPriorityType(row.priority)" size="small">
                {{ getPriorityLabel(row.priority) }}优先级
              </el-tag>
              <el-tag v-if="row.ticket_type" type="info" size="small">
                {{ row.ticket_type }}
              </el-tag>
              <el-tag v-if="row.project" type="warning" size="small">
                {{ row.project }}
              </el-tag>
              <el-tag v-if="!row.priority && !row.ticket_type && !row.project" type="success" size="small">
                默认规则
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="普通时限" width="120" align="center">
          <template #default="{ row }">
            <span class="sla-time">{{ formatDuration(row.normal_limit) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="严重时限" width="120" align="center">
          <template #default="{ row }">
            <span class="sla-time severe">{{ formatDuration(row.severe_limit) }}</span>
          </template>
        </el-table-column>

        <el-table-column label="匹配优先级" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.priority_order }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-switch
              v-model="row.is_active"
              @change="handleToggle(row)"
              active-text="启用"
              inactive-text="禁用"
              class="small-switch"
            />
          </template>
        </el-table-column>

        <el-table-column label="操作" width="150" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingRule ? '编辑规则' : '新建规则'"
      width="560px"
      class="modern-dialog"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules2"
        label-width="120px"
        label-position="top"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入规则名称" />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="适用优先级">
              <el-select v-model="form.priority" placeholder="全部优先级" clearable>
                <el-option label="低" value="low" />
                <el-option label="中" value="medium" />
                <el-option label="高" value="high" />
                <el-option label="紧急" value="critical" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工单类型">
              <el-input v-model="form.ticket_type" placeholder="如：bug、feature" clearable />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="项目标识">
          <el-input v-model="form.project" placeholder="留空表示适用所有项目" clearable />
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="普通时限(小时)" prop="normal_limit_hours">
              <el-input-number v-model="form.normal_limit_hours" :min="1" :max="720" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="严重时限(小时)" prop="severe_limit_hours">
              <el-input-number v-model="form.severe_limit_hours" :min="1" :max="720" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-alert
          title="规则匹配说明"
          type="info"
          :closable="false"
          show-icon
        >
          <template #default>
            规则越具体（指定的条件越多），匹配优先级越高。系统会优先使用最匹配的规则。
          </template>
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ editingRule ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getSLARules,
  createSLARule,
  updateSLARule,
  toggleSLARule,
  deleteSLARule
} from '@/api'

const loading = ref(false)
const submitting = ref(false)
const rules = ref([])
const showCreateDialog = ref(false)
const editingRule = ref(null)
const formRef = ref(null)

const form = reactive({
  name: '',
  priority: '',
  ticket_type: '',
  project: '',
  normal_limit_hours: 24,
  severe_limit_hours: 48
})

const rules2 = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  normal_limit_hours: [{ required: true, message: '请输入普通时限', trigger: 'blur' }],
  severe_limit_hours: [{ required: true, message: '请输入严重时限', trigger: 'blur' }]
}

const fetchRules = async () => {
  loading.value = true
  try {
    const data = await getSLARules()
    rules.value = data || []
  } catch (error) {
    console.error('获取时限规则失败:', error)
  } finally {
    loading.value = false
  }
}

const handleToggle = async (row) => {
  try {
    await toggleSLARule(row.id, row.is_active)
    ElMessage.success(row.is_active ? '已启用' : '已禁用')
  } catch (error) {
    row.is_active = !row.is_active
    ElMessage.error('操作失败')
  }
}

const handleEdit = (row) => {
  editingRule.value = row
  Object.assign(form, {
    name: row.name,
    priority: row.priority || '',
    ticket_type: row.ticket_type || '',
    project: row.project || '',
    // 将纳秒转换为小时（1小时 = 3.6e12纳秒）
    normal_limit_hours: Math.round(row.normal_limit / 3600000000000),
    severe_limit_hours: Math.round(row.severe_limit / 3600000000000)
  })
  showCreateDialog.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
      type: 'warning'
    })
    await deleteSLARule(row.id)
    ElMessage.success('删除成功')
    fetchRules()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    if (form.severe_limit_hours <= form.normal_limit_hours) {
      ElMessage.warning('严重时限必须大于普通时限')
      return
    }

    submitting.value = true
    const data = {
      name: form.name,
      priority: form.priority || '',
      ticket_type: form.ticket_type || '',
      project: form.project || '',
      normal_limit_hours: form.normal_limit_hours,
      severe_limit_hours: form.severe_limit_hours
    }

    if (editingRule.value) {
      await updateSLARule(editingRule.value.id, data)
      ElMessage.success('更新成功')
    } else {
      await createSLARule(data)
      ElMessage.success('创建成功')
    }

    showCreateDialog.value = false
    resetForm()
    fetchRules()
  } catch (error) {
    if (error !== false) {
      ElMessage.error('操作失败')
    }
  } finally {
    submitting.value = false
  }
}

const resetForm = () => {
  editingRule.value = null
  Object.assign(form, {
    name: '',
    priority: '',
    ticket_type: '',
    project: '',
    normal_limit_hours: 24,
    severe_limit_hours: 48
  })
}

const formatDuration = (nanoseconds) => {
  if (!nanoseconds) return '-'
  // 将纳秒转换为小时（1小时 = 3.6e12纳秒）
  const hours = Math.round(nanoseconds / 3600000000000)
  if (hours >= 24) {
    const days = Math.floor(hours / 24)
    const remainHours = hours % 24
    return remainHours > 0 ? `${days}天${remainHours}小时` : `${days}天`
  }
  return `${hours}小时`
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
  fetchRules()
})
</script>

<style scoped>
.sla-rules-container {
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

.scope-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.sla-time {
  font-weight: 600;
  color: var(--text-primary);
}

.sla-time.severe {
  color: var(--danger-color);
}

.modern-dialog :deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-light);
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.modern-dialog :deep(.el-select),
.modern-dialog :deep(.el-input-number) {
  width: 100%;
}

.small-switch :deep(.el-switch__label) {
  font-size: 12px;
  line-height: 1;
  transform: translateY(-2px);
}
</style>
