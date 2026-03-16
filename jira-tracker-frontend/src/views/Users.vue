<template>
  <div class="users-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1>用户管理</h1>
        <p>管理系统用户和权限</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
      </div>
    </div>

    <!-- 用户卡片网格 -->
    <div class="users-grid">
      <div
        v-for="user in users"
        :key="user.id"
        class="user-card"
        :class="{ disabled: user.status === 0 }"
      >
        <div class="user-header">
          <el-avatar :size="56" class="user-avatar">
            {{ user.name?.charAt(0) }}
          </el-avatar>
          <div class="user-basic">
            <h3>{{ user.name }}</h3>
            <el-tag :type="getRoleType(user.role)" size="small" effect="dark">
              {{ getRoleLabel(user.role) }}
            </el-tag>
          </div>
          <el-dropdown @command="(cmd) => handleCommand(cmd, user)" trigger="click">
            <el-button circle>
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-dropdown-item>
                <el-dropdown-item command="toggle">
                  <el-icon><SwitchButton /></el-icon>
                  {{ user.status === 1 ? '禁用' : '启用' }}
                </el-dropdown-item>
                <el-dropdown-item command="delete" divided>
                  <el-icon><Delete /></el-icon>
                  删除
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        
        <div class="user-stats">
          <div class="stat">
            <el-icon><Phone /></el-icon>
            <span>{{ user.phone }}</span>
          </div>
          <div class="stat">
            <el-icon><Document /></el-icon>
            <span>{{ user.ticket_num }} 个待处理</span>
          </div>
        </div>
        
        <div class="user-footer">
          <span class="status-dot" :class="user.status === 1 ? 'active' : 'inactive'"></span>
          <span class="status-text">{{ user.status === 1 ? '已启用' : '已禁用' }}</span>
          <span class="join-time">{{ formatTime(user.created_at) }} 加入</span>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-container" v-if="pagination.total > 0">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[12, 24, 48]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchUsers"
        @current-change="fetchUsers"
      />
    </div>

    <!-- 创建用户对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="新增用户"
      width="480px"
      class="modern-dialog"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="createForm.phone" placeholder="请输入手机号">
            <template #prefix>
              <el-icon><Phone /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="姓名" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入姓名">
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin">
              <el-tag type="danger" size="small">管理员</el-tag>
            </el-option>
            <el-option label="测试" value="test">
              <el-tag type="warning" size="small">测试</el-tag>
            </el-option>
            <el-option label="研发" value="dev">
              <el-tag type="primary" size="small">研发</el-tag>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="createLoading">
          创建用户
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑用户对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑用户"
      width="480px"
      class="modern-dialog"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="80px"
      >
        <el-form-item label="姓名" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入姓名" />
        </el-form-item>
        
        <el-form-item label="角色" prop="role">
          <el-select v-model="editForm.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin" />
            <el-option label="测试" value="test" />
            <el-option label="研发" value="dev" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="editForm.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleUpdate" :loading="editLoading">
          更新用户
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUsers, createUser, updateUser, deleteUser } from '@/api'
import dayjs from 'dayjs'

const users = ref([])
const loading = ref(false)
const createLoading = ref(false)
const editLoading = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const createFormRef = ref(null)
const editFormRef = ref(null)

const pagination = reactive({
  page: 1,
  pageSize: 12,
  total: 0
})

const createForm = reactive({
  phone: '',
  name: '',
  role: ''
})

const editForm = reactive({
  id: null,
  name: '',
  role: '',
  status: 1
})

const createRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

const editRules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

const fetchUsers = async () => {
  try {
    loading.value = true
    const data = await getUsers({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    users.value = data.list
    pagination.total = data.total
  } catch (error) {
    console.error('获取用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleCommand = async (command, user) => {
  if (command === 'edit') {
    Object.assign(editForm, {
      id: user.id,
      name: user.name,
      role: user.role,
      status: user.status
    })
    showEditDialog.value = true
  } else if (command === 'toggle') {
    const newStatus = user.status === 1 ? 0 : 1
    const action = newStatus === 1 ? '启用' : '禁用'
    
    try {
      await ElMessageBox.confirm(`确定要${action}用户 ${user.name} 吗？`, '提示', {
        type: 'warning'
      })
      await updateUser(user.id, { status: newStatus })
      ElMessage.success(`${action}成功`)
      fetchUsers()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('操作失败:', error)
      }
    }
  } else if (command === 'delete') {
    try {
      await ElMessageBox.confirm(`确定要删除用户 ${user.name} 吗？`, '提示', {
        type: 'warning'
      })
      await deleteUser(user.id)
      ElMessage.success('删除成功')
      fetchUsers()
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除用户失败:', error)
      }
    }
  }
}

const handleCreate = async () => {
  try {
    await createFormRef.value.validate()
    createLoading.value = true
    
    await createUser(createForm)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    
    Object.assign(createForm, {
      phone: '',
      name: '',
      role: ''
    })
    
    fetchUsers()
  } catch (error) {
    if (error !== false) {
      console.error('创建用户失败:', error)
    }
  } finally {
    createLoading.value = false
  }
}

const handleUpdate = async () => {
  try {
    await editFormRef.value.validate()
    editLoading.value = true
    
    await updateUser(editForm.id, {
      name: editForm.name,
      role: editForm.role,
      status: editForm.status
    })
    
    ElMessage.success('更新成功')
    showEditDialog.value = false
    fetchUsers()
  } catch (error) {
    if (error !== false) {
      console.error('更新用户失败:', error)
    }
  } finally {
    editLoading.value = false
  }
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD')
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
  fetchUsers()
})
</script>

<style scoped>
.users-container {
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

.users-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.user-card {
  background: var(--bg-primary);
  border-radius: 16px;
  padding: 20px;
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border-light);
  transition: all var(--transition-base);
}

.user-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.user-card.disabled {
  opacity: 0.6;
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
  font-size: 20px;
}

.user-basic {
  flex: 1;
}

.user-basic h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.user-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 8px;
}

.stat {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

.user-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--border-light);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.active {
  background: var(--success-color);
}

.status-dot.inactive {
  background: var(--text-disabled);
}

.status-text {
  font-size: 13px;
  color: var(--text-secondary);
  flex: 1;
}

.join-time {
  font-size: 12px;
  color: var(--text-tertiary);
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  background: var(--bg-primary);
  padding: 16px 20px;
  border-radius: 12px;
  border: 1px solid var(--border-light);
}

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
</style>
