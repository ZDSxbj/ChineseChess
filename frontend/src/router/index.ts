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
