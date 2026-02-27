import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000
})

// 防止重复显示错误消息
let messageQueue = []
let isShowingMessage = false

const showErrorMessage = (msg) => {
  // 如果已经在显示消息，则加入队列
  if (isShowingMessage) {
    // 避免重复消息
    if (!messageQueue.includes(msg)) {
      messageQueue.push(msg)
    }
    return
  }

  isShowingMessage = true
  ElMessage({
    message: msg,
    type: 'error',
    duration: 3000,
    onClose: () => {
      isShowingMessage = false
      // 显示队列中的下一条消息
      if (messageQueue.length > 0) {
        const nextMsg = messageQueue.shift()
        showErrorMessage(nextMsg)
      }
    }
  })
}

// 请求拦截器
request.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    console.log('发送请求:', config.method?.toUpperCase(), config.url)
    console.log('请求头:', config.headers)
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    // 如果配置了跳过错误处理，则不显示消息
    if (error.config?.skipErrorHandler) {
      return Promise.reject(error)
    }

    if (error.response) {
      const errorMsg = error.response.data?.error || '请求失败'
      switch (error.response.status) {
        case 401:
          showErrorMessage('登录已过期，请重新登录')
          const userStore = useUserStore()
          userStore.logout()
          router.push('/login')
          break
        default:
          showErrorMessage(errorMsg)
      }
    } else {
      showErrorMessage('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export default request
