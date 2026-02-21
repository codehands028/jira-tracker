import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '数据看板' }
      },
      {
        path: 'tickets',
        name: 'Tickets',
        component: () => import('@/views/Tickets.vue'),
        meta: { title: '工单列表' }
      },
      {
        path: 'tickets/:id',
        name: 'TicketDetail',
        component: () => import('@/views/TicketDetail.vue'),
        meta: { title: '工单详情' }
      },
      {
        path: 'statistics',
        name: 'Statistics',
        component: () => import('@/views/Statistics.vue'),
        meta: { title: '数据统计', requiresAdmin: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', requiresAdmin: true }
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('@/views/Notifications.vue'),
        meta: { title: '通知中心' }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/Profile.vue'),
        meta: { title: '个人信息' }
      },
      {
        path: 'logs',
        name: 'OperationLogs',
        component: () => import('@/views/OperationLogs.vue'),
        meta: { title: '操作日志', requiresAdmin: true }
      },
      {
        path: 'timeout-rules',
        name: 'TimeoutRules',
        component: () => import('@/views/TimeoutRules.vue'),
        meta: { title: '超时规则', requiresAdmin: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.token) {
    next('/login')
  } else if (to.meta.requiresAdmin && userStore.role !== 'admin') {
    next('/')
  } else if (to.path === '/login' && userStore.token) {
    next('/')
  } else {
    next()
  }
})

export default router
