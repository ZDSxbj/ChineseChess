import { defineStore } from 'pinia'

import { computed, ref, unref } from 'vue'

import apiBus from '@/utils/apiBus'
// 管理token
export const useUserStore = defineStore('user', () => {
  const t = ref<string>('')
  const userInfo = ref<UserInfo>()

  const token = computed(() => {
    if (t.value) {
      return t.value
    }
    else if (localStorage.getItem('token')) {
      t.value = localStorage.getItem('token') as string
      return t.value
    }
    apiBus.emit('API:UN_AUTH', null)
    return ''
  })

  const logout = () => {
    t.value = ''
    userInfo.value = undefined
  }

  const setToken = (newToken: string) => {
    const tokenValue = unref(newToken)
    t.value = tokenValue
    localStorage.setItem('token', tokenValue)
  }

  const setUser = (user: UserInfo) => {
    const newUser = unref(user)
    userInfo.value = newUser
  }

  apiBus.on('API:LOGIN', (req) => {
    const { token: newToken, name, avatar, exp } = req
    setToken(newToken)
    setUser({ name, avatar, exp })
  })

  apiBus.on('API:UN_AUTH', () => {
    logout()
  })

  apiBus.on('API:LOGOUT', () => {
    logout()
  })

  return { token, userInfo }
})
