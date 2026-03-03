
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getNotifications } from '@/api'

export const useNotificationStore = defineStore('notification', () => {
  const unreadCount = ref(0)

  const fetchNotifications = async () => {
    try {
      const data = await getNotifications({ page: 1, page_size: 100 })
      unreadCount.value = data.list.filter(n => !n.is_read).length
    } catch (error) {
      // 静默失败，不影响用户体验
      console.error('获取通知失败:', error)
    }
  }

  const decrementUnreadCount = () => {
    if (unreadCount.value > 0) {
      unreadCount.value--
    }
  }

  return {
    unreadCount,
    fetchNotifications,
    decrementUnreadCount
  }
})
