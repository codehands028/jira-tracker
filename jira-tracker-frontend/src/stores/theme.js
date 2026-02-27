import { ref, watch, onUnmounted } from 'vue'

// 主题模式：light | dark | system
const THEME_KEY = 'theme'

// 检查是否在浏览器环境
const isBrowser = () => typeof window !== 'undefined'

// 获取系统主题偏好
const getSystemTheme = () => {
  if (isBrowser() && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'light'
}

// 从localStorage获取保存的主题，默认使用浅色模式
const getSavedTheme = () => {
  if (isBrowser()) {
    return localStorage.getItem(THEME_KEY) || 'light'
  }
  return 'light'
}

const themeMode = ref('system') // 用户设置的模式：light | dark | system
const currentTheme = ref('light') // 当前实际应用的主题：light | dark

// 应用主题到document
const applyTheme = (theme) => {
  if (!isBrowser()) return
  
  if (theme === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
    document.body.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.removeAttribute('data-theme')
    document.body.removeAttribute('data-theme')
  }
}

// 根据用户设置计算实际主题
const resolveTheme = (mode) => {
  if (mode === 'system') {
    return getSystemTheme()
  }
  return mode
}

// 监听系统主题变化
let mediaQuery = null
const handleSystemThemeChange = (e) => {
  if (themeMode.value === 'system') {
    currentTheme.value = e.matches ? 'dark' : 'light'
    applyTheme(currentTheme.value)
  }
}

// 初始化系统主题监听
const initSystemThemeListener = () => {
  if (isBrowser() && window.matchMedia) {
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', handleSystemThemeChange)
  }
}

// 清理监听
const cleanupSystemThemeListener = () => {
  if (mediaQuery) {
    mediaQuery.removeEventListener('change', handleSystemThemeChange)
    mediaQuery = null
  }
}

// 初始化主题（在浏览器环境中调用）
const initTheme = () => {
  if (!isBrowser()) return
  
  const savedTheme = getSavedTheme()
  themeMode.value = savedTheme
  currentTheme.value = resolveTheme(savedTheme)
  applyTheme(currentTheme.value)
  initSystemThemeListener()
}

// 监听主题模式变化
watch(themeMode, (newMode) => {
  if (!isBrowser()) return
  
  localStorage.setItem(THEME_KEY, newMode)
  currentTheme.value = resolveTheme(newMode)
  applyTheme(currentTheme.value)
})

// 初始化
initTheme()

// 切换主题（循环切换：light -> dark -> system -> light）
export const toggleTheme = () => {
  const modes = ['light', 'dark', 'system']
  const currentIndex = modes.indexOf(themeMode.value)
  const nextIndex = (currentIndex + 1) % modes.length
  themeMode.value = modes[nextIndex]
}

// 切换明暗主题（简单切换：light <-> dark）
export const toggleDarkMode = () => {
  if (currentTheme.value === 'dark') {
    themeMode.value = 'light'
  } else {
    themeMode.value = 'dark'
  }
}

// 获取当前主题模式设置
export const getThemeMode = () => themeMode.value

// 获取当前实际应用的主题
export const getTheme = () => currentTheme.value

// 设置主题
export const setTheme = (newTheme) => {
  if (['light', 'dark', 'system'].includes(newTheme)) {
    themeMode.value = newTheme
  }
}

// 导出生命周期钩子，用于组件中清理
export const useThemeCleanup = () => {
  onUnmounted(() => {
    cleanupSystemThemeListener()
  })
}
