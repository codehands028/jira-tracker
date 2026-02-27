<template>
  <div class="profile-container">
    <el-card class="profile-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">个人信息</span>
        </div>
      </template>

      <el-form 
        ref="profileFormRef"
        :model="profileForm"
        :rules="rules"
        label-width="100px"
        class="profile-form"
      >
        <el-form-item label="头像">
          <el-avatar :size="80" class="profile-avatar">
            {{ profileForm.name?.charAt(0) || 'U' }}
          </el-avatar>
        </el-form-item>

        <el-form-item label="手机号">
          <el-input v-model="profileForm.phone" disabled />
        </el-form-item>

        <el-form-item label="姓名" prop="name">
          <el-input 
            v-model="profileForm.name" 
            placeholder="请输入姓名"
            maxlength="20"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="角色">
          <el-tag :type="getRoleTagType(profileForm.role)" effect="plain">
            {{ getRoleLabel(profileForm.role) }}
          </el-tag>
        </el-form-item>

        <el-form-item label="状态">
          <el-tag :type="profileForm.status === 1 ? 'success' : 'danger'" effect="plain">
            {{ profileForm.status === 1 ? '正常' : '禁用' }}
          </el-tag>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleUpdate" :loading="loading">
            保存修改
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getProfile, updateProfile } from '@/api'
import { useUserStore } from '@/stores/user'

const profileFormRef = ref(null)
const loading = ref(false)
const userStore = useUserStore()

const profileForm = reactive({
  id: '',
  name: '',
  phone: '',
  role: '',
  status: 1
})

const rules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
  ]
}

const getRoleLabel = (role) => {
  const map = {
    admin: '管理员',
    test: '测试',
    dev: '研发'
  }
  return map[role] || '未知角色'
}

const getRoleTagType = (role) => {
  const map = {
    admin: 'danger',
    test: 'warning',
    dev: 'primary'
  }
  return map[role] || 'info'
}

const fetchProfile = async () => {
  try {
    const { data } = await getProfile()
    Object.assign(profileForm, data)
  } catch (error) {
    console.error('获取个人信息失败:', error)
  }
}

const handleUpdate = async () => {
  if (!profileFormRef.value) return

  await profileFormRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      await updateProfile({ name: profileForm.name })
      ElMessage.success('更新成功')
      // 重新获取个人信息
      await fetchProfile()
      // 同步更新userStore中的用户信息
      userStore.setUserInfo(profileForm)
    } catch (error) {
      console.error('更新个人信息失败:', error)
    } finally {
      loading.value = false
    }
  })
}

onMounted(() => {
  fetchProfile()
})
</script>

<style scoped>
.profile-container {
  max-width: 800px;
  margin: 20px auto;
  padding: 0 20px;
}

.profile-card {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.profile-form {
  padding: 20px 40px;
}

.profile-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  font-size: 32px;
  font-weight: bold;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input.is-disabled .el-input__inner) {
  color: #909399;
}

/* 深色模式样式 */
[data-theme="dark"] .profile-card {
  background-color: var(--bg-primary);
  border-color: var(--border-color);
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.3);
}

[data-theme="dark"] .card-title {
  color: var(--text-primary);
}

[data-theme="dark"] :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

[data-theme="dark"] :deep(.el-input__wrapper) {
  background-color: var(--bg-secondary);
  box-shadow: 0 0 0 1px var(--border-color) inset;
}

[data-theme="dark"] :deep(.el-input__inner) {
  color: var(--text-primary);
}

[data-theme="dark"] :deep(.el-input.is-disabled .el-input__inner) {
  color: var(--text-disabled);
}

[data-theme="dark"] :deep(.el-tag--plain) {
  background-color: var(--bg-secondary);
  border-color: var(--border-color);
  color: var(--text-primary);
}

/* 头像标签垂直居中 */
:deep(.el-form-item:first-child) {
  align-items: flex-start;
  padding-top: 20px;
}

:deep(.el-form-item:first-child .el-form-item__label) {
  line-height: 40px;
  margin-top: 10px;
}

/* 手机号输入框深色模式 */
[data-theme="dark"] :deep(.el-input.is-disabled .el-input__wrapper) {
  background-color: var(--bg-tertiary) !important;
  box-shadow: 0 0 0 1px var(--border-color) inset !important;
}

[data-theme="dark"] :deep(.el-input.is-disabled .el-input__inner) {
  color: var(--text-disabled) !important;
  -webkit-text-fill-color: var(--text-disabled) !important;
}

/* 输入框字数统计深色模式 */
[data-theme="dark"] :deep(.el-input__count) {
  color: var(--text-secondary) !important;
  background: transparent !important;
}

[data-theme="dark"] :deep(.el-input__count-inner) {
  color: var(--text-secondary) !important;
  background: transparent !important;
}
</style>
