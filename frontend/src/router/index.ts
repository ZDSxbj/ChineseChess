import { createRouter, createWebHistory } from 'vue-router'

import apiBus from '@/utils/apiBus'
import channel from '@/utils/channel'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('../layout/DefaultLayout.vue'),
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('../views/default/HomeView.vue'),
        },
        {
          path: 'about',
          name: 'about',
          component: () => import('../views/default/AboutView.vue'),
        },
        {
          path: 'room',
          name: 'room',
          component: () => import('../views/default/RoomView.vue'),
        },
        // 个人主页路由（带登录校验）
        { 
          path: 'profile', 
          name: 'profile', 
          component: () => import('../views/profile/ProfileHub.vue'), 
          meta: { requiresAuth: true }, // 必须登录才能访问
          children: [
            {
              path: 'info', // 子路由，个人信息
              name: 'profile-info',
              component: () => import('../views/profile/Profile.vue') // 指向个人信息详情页
            },
            {
              path: 'records', // 子路由，行棋记录
              name: 'profile-records',
              component: () => import('../views/profile/GameRecords.vue') // 指向行棋记录页
            }
          ]
        }
      ],
    },
    {
      path: '/auth/',
      name: 'auth',
      component: () => import('../layout/AuthLayout.vue'),
      children: [
        {
          path: 'login',
          name: 'login',
          component: () => import('../views/auth/LoginView.vue'),
        },
        {
          path: 'register',
          name: 'register',
          component: () => import('../views/auth/RegisterView.vue'),
        },
      ],
    },
    {
      path: '/game/',
      name: 'game',
      component: () => import('../layout/GameLayout.vue'),
      children: [
        {
          path: 'chess',
          name: 'game-chess',
          component: () => import('../views/game/ChessView.vue'),
        },
        {
          path: 'replay',
          name: 'game-replay',
          component: () => import('../views/game/ReplayView.vue'),
        },
      ],
    },
  ],
})

// 路由守卫：校验需要登录的路由
router.beforeEach((to, from, next) => {
  const requiresAuth = to.meta.requiresAuth
  // 根据实际项目调整登录判断（如Pinia状态、localStorage token）
  const isLoggedIn = localStorage.getItem('token') || false

  if (requiresAuth && !isLoggedIn) {
    next('/auth/login') // 未登录则跳登录页
  } else {
    next() // 已登录或无需登录，正常跳转
  }
})

function logout() {
  router.push('/auth/login')
}

apiBus.on('API:UN_AUTH', () => {
  logout()
})

apiBus.on('API:LOGOUT', () => {
  logout()
})

apiBus.on('API:LOGIN', (req) => {
  const { stop } = req
  if (stop)
    return
  router.push('/')
})

channel.on('MATCH:SUCCESS', () => {
  router.push('/game/chess')
})

export default router
