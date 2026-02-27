<template>
  <div class="login-container">
    <!-- 背景动画 -->
    <div class="background-animation">
      <div class="floating-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
        <div class="shape shape-3"></div>
        <div class="shape shape-4"></div>
        <div class="shape shape-5"></div>
      </div>
    </div>
    
    <!-- 登录卡片 -->
    <div class="login-wrapper">
      <div class="login-card">
        <!-- 左侧装饰 -->
        <div class="login-decoration">
          <div class="decoration-content">
            <div class="icon-wrapper">
              <el-icon class="main-icon"><Document /></el-icon>
            </div>
            <h1>Jira Tracker</h1>
            <p>工单跟踪与效率管理平台</p>
            <div class="features">
              <div class="feature-item">
                <el-icon><Check /></el-icon>
                <span>工单流转管理</span>
              </div>
              <div class="feature-item">
                <el-icon><Check /></el-icon>
                <span>自动超时提醒</span>
              </div>
              <div class="feature-item">
                <el-icon><Check /></el-icon>
                <span>数据统计分析</span>
              </div>
            </div>
          </div>
          <div class="decoration-shapes">
            <div class="circle circle-1"></div>
            <div class="circle circle-2"></div>
            <div class="circle circle-3"></div>
          </div>
        </div>
        
        <!-- 右侧登录表单 -->
        <div class="login-form-container">
          <div class="form-header">
            <h2>欢迎登录</h2>
            <p>请输入手机号和验证码登录系统</p>
          </div>
          
          <el-form
            ref="loginFormRef"
            :model="loginForm"
            :rules="loginRules"
            class="login-form"
            size="large"
          >
            <el-form-item prop="phone">
              <el-input
                v-model="loginForm.phone"
                placeholder="请输入手机号"
                :prefix-icon="Phone"
                clearable
              />
            </el-form-item>
            
            <el-form-item prop="code">
              <el-input
                v-model="loginForm.code"
                placeholder="请输入验证码"
                :prefix-icon="Key"
                maxlength="6"
                @keyup.enter="handleLogin"
              >
                <template #append>
                  <el-button
                    :disabled="countdown > 0"
                    :loading="sendingCode"
                    @click="handleSendCode"
                    class="code-btn"
                  >
                    {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
                  </el-button>
                </template>
              </el-input>
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                :loading="loading"
                @click="handleLogin"
                class="login-btn"
              >
                <span v-if="!loading">登 录</span>
                <span v-else>登录中...</span>
              </el-button>
            </el-form-item>
          </el-form>
          
          <div class="login-tips">
            <el-icon><InfoFilled /></el-icon>
            <span>开发环境验证码可输入任意6位数字</span>
          </div>
          
          <div class="test-accounts">
            <div class="test-title">测试账号：</div>
            <div class="account-list">
              <div class="account-item" @click="fillAccount('13800138000')">
                <span class="account-phone">13800138000</span>
                <el-tag type="danger" size="small">管理员</el-tag>
              </div>
              <div class="account-item" @click="fillAccount('13800138001')">
                <span class="account-phone">13800138001</span>
                <el-tag type="warning" size="small">测试</el-tag>
              </div>
              <div class="account-item" @click="fillAccount('13800138002')">
                <span class="account-phone">13800138002</span>
                <el-tag type="primary" size="small">研发</el-tag>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 底部信息 -->
      <div class="login-footer">
        <p>© 2024 Jira Tracker. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Phone, Key } from '@element-plus/icons-vue'
import { sendCode, login } from '@/api'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const loginFormRef = ref(null)
const loading = ref(false)
const sendingCode = ref(false)
const countdown = ref(0)

const loginForm = reactive({
  phone: '',
  code: ''
})

const loginRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码为6位数字', trigger: 'blur' }
  ]
}

const handleSendCode = async () => {
  try {
    await loginFormRef.value.validateField('phone')
    
    sendingCode.value = true
    await sendCode(loginForm.phone)
    ElMessage.success('验证码已发送')
    
    // 开始倒计时
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (error) {
    if (error !== false) {
      console.error('发送验证码失败:', error)
    }
  } finally {
    sendingCode.value = false
  }
}

const fillAccount = (phone) => {
  loginForm.phone = phone
}

const handleLogin = async () => {
  try {
    await loginFormRef.value.validate()
    
    loading.value = true
    const data = await login(loginForm.phone, loginForm.code)
    
    // 保存用户信息
    userStore.setToken(data.token)
    userStore.setUserInfo(data.user)
    
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error) {
    console.error('登录失败:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

/* 背景动画 */
.background-animation {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
}

.floating-shapes {
  position: absolute;
  width: 100%;
  height: 100%;
}

.shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.1;
  background: white;
  animation: float 20s infinite ease-in-out;
}

.shape-1 {
  width: 400px;
  height: 400px;
  top: -200px;
  left: -100px;
  animation-delay: 0s;
}

.shape-2 {
  width: 300px;
  height: 300px;
  top: 50%;
  right: -150px;
  animation-delay: 3s;
}

.shape-3 {
  width: 200px;
  height: 200px;
  bottom: -100px;
  left: 30%;
  animation-delay: 6s;
}

.shape-4 {
  width: 150px;
  height: 150px;
  top: 30%;
  left: 10%;
  animation-delay: 9s;
}

.shape-5 {
  width: 250px;
  height: 250px;
  bottom: 10%;
  right: 20%;
  animation-delay: 12s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) rotate(0deg);
  }
  25% {
    transform: translate(50px, 50px) rotate(90deg);
  }
  50% {
    transform: translate(0, 100px) rotate(180deg);
  }
  75% {
    transform: translate(-50px, 50px) rotate(270deg);
  }
}

/* 登录包装器 */
.login-wrapper {
  width: 100%;
  max-width: 900px;
  padding: 20px;
  z-index: 1;
}

/* 登录卡片 */
.login-card {
  display: flex;
  background: var(--bg-primary);
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  min-height: 500px;
}

/* 左侧装饰 */
.login-decoration {
  flex: 1;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 60px 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.decoration-content {
  position: relative;
  z-index: 2;
  color: white;
}

.icon-wrapper {
  width: 80px;
  height: 80px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}

.main-icon {
  font-size: 40px;
  color: white;
}

.decoration-content h1 {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 12px;
}

.decoration-content > p {
  font-size: 16px;
  opacity: 0.9;
  margin-bottom: 40px;
}

.features {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  opacity: 0.95;
}

.feature-item .el-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.decoration-shapes {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1;
}

.circle {
  position: absolute;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.circle-1 {
  width: 200px;
  height: 200px;
  top: -50px;
  right: -50px;
}

.circle-2 {
  width: 150px;
  height: 150px;
  bottom: -30px;
  left: -30px;
}

.circle-3 {
  width: 100px;
  height: 100px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

/* 右侧登录表单 */
.login-form-container {
  flex: 1;
  padding: 60px 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.form-header {
  margin-bottom: 32px;
}

.form-header h2 {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.form-header p {
  color: var(--text-tertiary);
  font-size: 14px;
}

.login-form {
  width: 100%;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 12px;
  padding: 4px 12px;
  height: 48px;
  box-shadow: none;
  border: 2px solid var(--border-color);
}

.login-form :deep(.el-input__wrapper:hover) {
  border-color: var(--primary-light);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.login-form :deep(.el-input__inner) {
  font-size: 15px;
}

.code-btn {
  background: transparent;
  border: none;
  color: var(--primary-color);
  font-weight: 600;
  padding: 0 16px;
}

.code-btn:hover {
  color: var(--primary-light);
}

.login-btn {
  width: 100%;
  height: 48px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  background: var(--primary-gradient);
  border: none;
  box-shadow: 0 4px 14px rgba(102, 126, 234, 0.4);
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

.login-tips {
  margin-top: 24px;
  padding: 12px 16px;
  background: rgba(102, 126, 234, 0.05);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-tertiary);
  font-size: 13px;
}

.login-tips .el-icon {
  color: var(--primary-color);
}

/* 测试账号 */
.test-accounts {
  margin-top: 20px;
  padding: 16px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.08) 100%);
  border-radius: 12px;
  border: 1px solid rgba(102, 126, 234, 0.15);
}

.test-title {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  font-weight: 500;
}

.account-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.account-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: var(--bg-secondary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.account-item:hover {
  border-color: var(--primary-color);
  transform: translateX(4px);
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}

.account-phone {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

/* 底部信息 */
.login-footer {
  text-align: center;
  margin-top: 32px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 13px;
}

/* 响应式 */
@media (max-width: 768px) {
  .login-card {
    flex-direction: column;
  }

  .login-decoration {
    padding: 40px 30px;
  }

  .decoration-content h1 {
    font-size: 24px;
  }

  .login-form-container {
    padding: 40px 30px;
  }

  .form-header h2 {
    font-size: 24px;
  }
}

/* 深色主题 */
[data-theme="dark"] .login-container {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

[data-theme="dark"] .login-card {
  background: var(--bg-primary);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

[data-theme="dark"] .login-decoration {
  background: linear-gradient(135deg, #1E88E5 0%, #1565C0 100%);
}

[data-theme="dark"] .login-form-container {
  background: var(--bg-primary);
}

[data-theme="dark"] .form-header h2 {
  color: var(--text-primary);
}

[data-theme="dark"] .form-header p {
  color: var(--text-tertiary);
}

[data-theme="dark"] .login-form :deep(.el-input__wrapper) {
  background: var(--bg-secondary);
  border-color: var(--border-color);
}

[data-theme="dark"] .login-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

[data-theme="dark"] .login-form :deep(.el-input__wrapper.is-focus) {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(30, 136, 229, 0.2);
}

[data-theme="dark"] .code-btn {
  color: var(--primary-color);
}

[data-theme="dark"] .login-btn {
  background: linear-gradient(135deg, #1E88E5 0%, #1565C0 100%);
  box-shadow: 0 4px 14px rgba(30, 136, 229, 0.4);
}

[data-theme="dark"] .login-btn:hover {
  box-shadow: 0 6px 20px rgba(30, 136, 229, 0.5);
}

[data-theme="dark"] .login-tips {
  background: rgba(30, 136, 229, 0.1);
}

[data-theme="dark"] .login-tips .el-icon {
  color: var(--primary-color);
}

[data-theme="dark"] .test-accounts {
  background: rgba(30, 136, 229, 0.08);
  border-color: rgba(30, 136, 229, 0.2);
}

[data-theme="dark"] .account-item {
  background: var(--bg-secondary);
}

[data-theme="dark"] .account-item:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px rgba(30, 136, 229, 0.2);
}

[data-theme="dark"] .account-phone {
  color: var(--text-primary);
}

[data-theme="dark"] .login-footer {
  color: rgba(255, 255, 255, 0.5);
}
</style>
