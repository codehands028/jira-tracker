import { ref, watch } from 'vue'

// 从localStorage获取保存的主题，默认为light
const savedTheme = localStorage.getItem('theme') || 'light'
const theme = ref(savedTheme)

// 监听主题变化，保存到localStorage并应用到document
watch(theme, (newTheme) => {
  localStorage.setItem('theme', newTheme)
  if (newTheme === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.removeAttribute('data-theme')
  }
}, { immediate: true })

// 切换主题
export const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
}

// 获取当前主题
export const getTheme = () => theme.value

// 设置主题
export const setTheme = (newTheme) => {
  if (newTheme === 'light' || newTheme === 'dark') {
    theme.value = newTheme
  }
}
