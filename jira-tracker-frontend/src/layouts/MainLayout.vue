<template>
  <el-container class="main-container">
    <!-- 侧边栏 -->
    <el-aside :width="sidebarWidth" class="modern-sidebar">
      <div class="sidebar-header">
        <div class="logo-container">
          <div class="logo-icon">
            <el-icon><Document /></el-icon>
          </div>
          <transition name="fade">
            <span v-show="!isCollapsed" class="logo-text">Jira Tracker</span>
          </transition>
        </div>
        <el-button
          class="collapse-btn"
          :icon="isCollapsed ? 'Expand' : 'Fold'"
          @click="toggleSidebar"
          circle
          size="small"
        />
      </div>
      
      <el-menu
        :default-active="activeMenu"
        class="modern-menu"
        :collapse="isCollapsed"
        :collapse-transition="false"
        @select="handleSelect"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataBoard /></el-icon>
          <template #title>数据看板</template>
        </el-menu-item>
        
        <el-menu-item index="/tickets">
          <el-icon><Tickets /></el-icon>
          <template #title>工单列表</template>
        </el-menu-item>
        
        <el-menu-item index="/notifications">
          <el-icon><Bell /></el-icon>
          <template #title>
            <div class="menu-item-content">
              <span>通知中心</span>
              <el-badge
                v-if="unreadCount > 0"
                :value="unreadCount"
                class="menu-badge"
              />
            </div>
          </template>
        </el-menu-item>
        
        <el-divider v-if="isAdmin" class="menu-divider" />
        
        <el-menu-item v-if="isAdmin" index="/statistics">
          <el-icon><DataAnalysis /></el-icon>
          <template #title>数据统计</template>
        </el-menu-item>
        
        <el-menu-item v-if="isAdmin" index="/users">
          <el-icon><User /></el-icon>
          <template #title>用户管理</template>
        </el-menu-item>
        
        <el-menu-item v-if="isAdmin" index="/logs">
          <el-icon><List /></el-icon>
          <template #title>操作日志</template>
        </el-menu-item>
        
        <el-menu-item v-if="isAdmin" index="/timeout-rules">
          <el-icon><Clock /></el-icon>
          <template #title>超时规则</template>
        </el-menu-item>
        
        <el-menu-item v-if="isAdmin" index="/sla-rules">
          <el-icon><Timer /></el-icon>
          <template #title>时限规则</template>
        </el-menu-item>
      </el-menu>
      
      <!-- 用户信息卡片 -->
      <div class="user-card">
        <el-avatar :size="40" class="user-avatar">
          {{ userInfo?.name?.charAt(0) || 'U' }}
        </el-avatar>
        <transition name="fade">
          <div v-show="!isCollapsed" class="user-info">
            <div class="user-name">{{ userInfo?.name || '未登录' }}</div>
            <div class="user-role">
              <el-tag :type="getRoleTagType(userInfo?.role)" size="small" effect="plain">
                {{ getRoleLabel(userInfo?.role) }}
              </el-tag>
            </div>
          </div>
        </transition>
        <transition name="fade">
          <el-dropdown v-show="!isCollapsed" @command="handleCommand" trigger="click" placement="top">
            <el-icon class="more-icon"><MoreFilled /></el-icon>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人信息
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </transition>
      </div>
    </el-aside>

    <!-- 主内容区 -->
    <el-container class="main-content">
      <!-- 顶栏 -->
      <el-header class="modern-header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <div class="header-right">
          <el-tooltip :content="isDarkMode ? '切换到亮色模式' : '切换到深色模式'" placement="bottom">
            <el-button circle class="header-btn theme-toggle-btn" @click="toggleDarkMode">
              <el-icon>
                <Sunny v-if="isDarkMode" />
                <Moon v-else />
              </el-icon>
            </el-button>
          </el-tooltip>

          <el-button circle class="header-btn" @click="router.push('/notifications')">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
              <el-icon><Bell /></el-icon>
            </el-badge>
          </el-button>

          <el-divider direction="vertical" />
          
          <el-dropdown @command="handleCommand" trigger="click">
            <div class="header-user">
              <el-avatar :size="32" class="user-avatar-sm">
                {{ userInfo?.name?.charAt(0) || 'U' }}
              </el-avatar>
              <div class="user-detail">
                <span class="user-name-sm">{{ userInfo?.name || '未登录' }}</span>
                <el-tag :type="getRoleTagType(userInfo?.role)" size="small" effect="plain">
                  {{ getRoleLabel(userInfo?.role) }}
                </el-tag>
              </div>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人信息
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="content-area">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useNotificationStore } from '@/stores/notification'
import { toggleDarkMode, getTheme, getThemeMode } from '@/stores/theme'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const notificationStore = useNotificationStore()

const isCollapsed = ref(false)
const activeMenu = computed(() => route.path)
const currentTitle = computed(() => route.meta.title || '首页')
const userInfo = computed(() => userStore.userInfo)
const isAdmin = computed(() => userStore.role === 'admin')
const unreadCount = computed(() => notificationStore.unreadCount)
const isDarkMode = computed(() => getTheme() === 'dark')

const sidebarWidth = computed(() => isCollapsed.value ? '64px' : '240px')

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
}

const handleSelect = (index) => {
  router.push(index)
}

const handleCommand = (command) => {
  if (command === 'logout') {
    userStore.logout()
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/profile')
  }
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

let timer = null

onMounted(() => {
  notificationStore.fetchNotifications()
  timer = setInterval(() => {
    notificationStore.fetchNotifications()
  }, 5 * 60 * 1000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.main-container {
  height: 100vh;
  overflow: hidden;
}

/* 侧边栏样式 */
.modern-sidebar {
  background: linear-gradient(180deg, #ffffff 0%, #fafbfc 100%);
  border-right: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  transition: width var(--transition-base);
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.02);
  position: relative;
}

/* 深色主题侧边栏 */
[data-theme="dark"] .modern-sidebar {
  background: linear-gradient(180deg, #1E1E1E 0%, #121212 100%);
  border-right: 1px solid var(--border-light);
}

/* 侧边栏收起时的样式 */
.modern-sidebar[style*="64px"] {
  box-shadow: none;
}

.modern-sidebar[style*="64px"] .sidebar-header {
  padding: 16px 12px;
  justify-content: center;
}

.modern-sidebar[style*="64px"] .logo-container {
  justify-content: center;
}

.modern-sidebar[style*="64px"] .logo-text {
  display: none;
}

.modern-sidebar[style*="64px"] .collapse-btn {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border-color: var(--border-light);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 深色主题收起按钮 */
[data-theme="dark"] .modern-sidebar[style*="64px"] .collapse-btn {
  background: var(--bg-primary);
  border-color: var(--border-color);
}

.modern-sidebar[style*="64px"] .collapse-btn:hover {
  background: var(--primary-color);
  color: white;
  border-color: var(--primary-color);
}

.modern-sidebar[style*="64px"] .logo-icon {
  width: 36px;
  height: 36px;
  font-size: 18px;
}

.sidebar-header {
  padding: 20px 16px;
  border-bottom: 1px solid var(--border-light);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo-container {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  overflow: hidden;
}

.logo-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: var(--primary-gradient);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  background: var(--primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  white-space: nowrap;
}

.collapse-btn {
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.collapse-btn:hover {
  background: var(--bg-hover);
  color: var(--primary-color);
  border-color: var(--primary-light);
}

/* 菜单样式 */
.modern-menu {
  flex: 1;
  border-right: none;
  background: transparent;
  padding: 8px;
  overflow-y: auto;
  overflow-x: hidden;
}

.modern-menu:not(.el-menu--collapse) {
  width: 100%;
}

/* 收起时的菜单样式 */
.el-menu--collapse .el-menu-item {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 !important;
  height: 48px;
  margin: 4px 0;
  border-radius: 8px;
  position: relative;
}

.el-menu--collapse .el-menu-item .el-icon {
  margin: 0 !important;
  font-size: 20px;
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}

.el-menu--collapse .el-menu-item.is-active {
  background: var(--primary-gradient);
}

.el-menu--collapse .el-menu-item.is-active .el-icon {
  color: white;
}

.menu-divider {
  margin: 8px 0;
}

.menu-item-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
}

.menu-badge {
  margin-left: 0;
  margin-top: 0;
  margin-bottom: 0;
  vertical-align: middle;
  display: inline-flex;
  align-items: center;
}

/* 收起时隐藏通知badge */
.modern-sidebar[style*="64px"] .menu-badge {
  display: none;
}

/* 用户卡片 */
.user-card {
  padding: 16px;
  border-top: 1px solid var(--border-light);
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.02) 0%, rgba(118, 75, 162, 0.02) 100%);
}

/* 深色主题用户卡片 */
[data-theme="dark"] .user-card {
  background: linear-gradient(135deg, rgba(30, 136, 229, 0.1) 0%, rgba(21, 101, 192, 0.1) 100%);
}

/* 收起时的用户卡片 */
.modern-sidebar[style*="64px"] .user-card {
  padding: 12px;
  justify-content: center;
}

.modern-sidebar[style*="64px"] .user-avatar {
  width: 36px;
  height: 36px;
  font-size: 14px;
}

.modern-sidebar[style*="64px"] .user-info,
.modern-sidebar[style*="64px"] .more-icon {
  display: none;
}

.user-avatar {
  flex-shrink: 0;
  background: var(--primary-gradient);
  color: white;
  font-weight: 600;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

.more-icon {
  cursor: pointer;
  color: var(--text-tertiary);
  transition: color var(--transition-fast);
  padding: 4px;
  border-radius: 4px;
}

.more-icon:hover {
  color: var(--primary-color);
  background: var(--bg-hover);
}

/* 主内容区 */
.main-content {
  background: var(--bg-secondary);
  display: flex;
  flex-direction: column;
}

/* 顶栏样式 */
.modern-header {
  background: white;
  border-bottom: 1px solid var(--border-light);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.02);
  z-index: 10;
}

/* 深色主题顶栏 */
[data-theme="dark"] .modern-header {
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

/* 深色主题面包屑 */
[data-theme="dark"] .el-breadcrumb__inner {
  color: var(--text-secondary);
}

[data-theme="dark"] .el-breadcrumb__inner:hover {
  color: var(--primary-color);
}

/* 深色主题用户信息区域 */
[data-theme="dark"] .header-user {
  background: var(--bg-secondary);
}

[data-theme="dark"] .header-user:hover {
  background: var(--bg-hover);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-btn {
  border: 1px solid var(--border-color);
  background: transparent;
  color: var(--text-secondary);
}

.header-btn:hover {
  background: var(--bg-hover);
  color: var(--primary-color);
  border-color: var(--primary-light);
}

/* 主题切换按钮 */
.theme-toggle-btn {
  transition: all var(--transition-base);
}

.theme-toggle-btn .el-icon {
  transition: transform var(--transition-base);
}

[data-theme="dark"] .theme-toggle-btn {
  color: #fbbf24;
  border-color: #fbbf24;
}

[data-theme="dark"] .theme-toggle-btn:hover {
  color: #fcd34d;
  border-color: #fcd34d;
  background: rgba(251, 191, 36, 0.1);
}

.header-user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 14px 6px 6px;
  border-radius: 24px;
  background: var(--bg-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.header-user:hover {
  background: var(--bg-hover);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.user-detail {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-avatar-sm {
  background: var(--primary-gradient);
  color: white;
  font-weight: 600;
  font-size: 12px;
}

.user-name-sm {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

/* 内容区 */
.content-area {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-fast);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .modern-sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    z-index: 100;
    transform: translateX(-100%);
    transition: transform var(--transition-base);
  }
  
  .modern-sidebar.show {
    transform: translateX(0);
  }
  
  .header-user .user-name-sm {
    display: none;
  }
}
</style>
