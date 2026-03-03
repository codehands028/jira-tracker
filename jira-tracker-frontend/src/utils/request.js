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
    // 添加CSRF token到请求头
    if (userStore.csrfToken) {
      config.headers['X-CSRF-Token'] = userStore.csrfToken
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
    // 如果响应头中有新的CSRF token，更新本地存储
    const newCsrfToken = response.headers['x-csrf-token']
    if (newCsrfToken) {
      const userStore = useUserStore()
      userStore.setCsrfToken(newCsrfToken)
    }
    return response.data
  },
  error => {
    // 如果配置了跳过错误处理，则不显示消息
    if (error.config?.skipErrorHandler) {
      return Promise.reject(error)
    }

    if (error.response) {
      const errorMsg = error.response.data?.error || '请求失败'
      const statusCode = error.response.status

      // 检查是否是重复工单错误
      if (errorMsg.includes('Duplicate entry') && errorMsg.includes('jira_key')) {
        showErrorMessage('该工单编号已存在，请检查后重新输入')
        return Promise.reject(error)
      }

      // 根据状态码提供友好的错误提示
      switch (statusCode) {
        case 400:
          showErrorMessage(errorMsg || '请求参数有误，请检查输入内容')
          break
        case 401:
          showErrorMessage('登录已过期，请重新登录')
          const userStore = useUserStore()
          userStore.logout()
          router.push('/login')
          break
        case 403:
          showErrorMessage('您没有权限执行此操作')
          break
        case 404:
          showErrorMessage('请求的资源不存在')
          break
        case 429:
          showErrorMessage('操作过于频繁，请稍后再试')
          break
        case 500:
          showErrorMessage('服务器内部错误，请稍后重试或联系管理员')
          break
        case 502:
        case 503:
        case 504:
          showErrorMessage('服务暂时不可用，请稍后重试')
          break
        default:
          showErrorMessage(errorMsg)
      }
    } else if (error.code === 'ECONNABORTED') {
      showErrorMessage('请求超时，请检查网络连接后重试')
    } else if (error.message === 'Network Error') {
      showErrorMessage('网络连接失败，请检查网络设置')
    } else {
      showErrorMessage('请求失败，请稍后重试')
    }
    return Promise.reject(error)
  }
)

export default request
