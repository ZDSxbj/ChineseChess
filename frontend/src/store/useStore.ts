import { defineStore } from 'pinia'

import { computed, ref, unref } from 'vue'

import { getProfile } from '@/api/user/getProfile'
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
    // 尝试获取完整的用户资料（包含 id），以便前端用于消息发送者比对
    getProfile().then((resp) => {
      const data = (resp && typeof resp === 'object' && 'data' in resp) ? (resp as any).data : resp
      if (data) {
        setUser(data)
        return
      }
      // 兜底设置部分信息
      setUser({ name, avatar, exp })
    }).catch(() => {
      setUser({ name, avatar, exp })
    })
  })

  apiBus.on('API:UN_AUTH', () => {
    logout()
  })

  apiBus.on('API:LOGOUT', () => {
    logout()
  })

  // 暴露必要的操作方法，便于在页面中调用（例如 Profile.vue 中更新头像等）
  return { token, userInfo, setUser, setToken, logout }
})
