import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || '{}'))
  const csrfToken = ref(localStorage.getItem('csrfToken') || '')

  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUserInfo = (info) => {
    userInfo.value = info
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  const setCsrfToken = (newCsrfToken) => {
    csrfToken.value = newCsrfToken
    localStorage.setItem('csrfToken', newCsrfToken)
  }

  const logout = () => {
    token.value = ''
    userInfo.value = {}
    csrfToken.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    localStorage.removeItem('csrfToken')
  }

  // 使用computed确保role能响应式更新
  const role = computed(() => userInfo.value?.role || '')

  return {
    token,
    userInfo,
    csrfToken,
    role,
    setToken,
    setUserInfo,
    setCsrfToken,
    logout
  }
})
