import { defineStore } from 'pinia'

import { computed, ref, unref } from 'vue'

import { getProfile } from '@/api/user/getProfile'
import apiBus from '@/utils/apiBus'
// 管理token
export const useUserStore = defineStore('user', () => {
  const t = ref<string>('')
  const userInfo = ref<UserInfo>()

  const token = computed(() => {
    if (t.value) return t.value
    // 根据 rememberMe 选择存储源：localStorage(记住我) 或 sessionStorage(不记住)
    const storage = (localStorage.getItem('rememberMe') === '1') ? localStorage : sessionStorage
    const persisted = storage.getItem('token')
    if (persisted) {
      t.value = persisted as string
      return t.value
    }
    apiBus.emit('API:UN_AUTH', null)
    return ''
  })

  const logout = () => {
    t.value = ''
    userInfo.value = undefined
    // 清理两种存储的 token
    try { localStorage.removeItem('token') } catch {}
    try { sessionStorage.removeItem('token') } catch {}
  }

  const setToken = (newToken: string) => {
    const tokenValue = unref(newToken)
    t.value = tokenValue
    const storage = (localStorage.getItem('rememberMe') === '1') ? localStorage : sessionStorage
    storage.setItem('token', tokenValue)
  }

  const setUser = (user: UserInfo) => {
    const newUser = unref(user)
    userInfo.value = newUser
  }

  apiBus.on('API:LOGIN', (req) => {
    const { token: newToken, name, avatar, exp, remember } = req as any
    // 更新 rememberMe 标记：选中为 '1'，未选中为 '0'；未提供则保持不变
    if (typeof remember !== 'undefined') {
      try { localStorage.setItem('rememberMe', remember ? '1' : '0') } catch {}
    }
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
    // 登出不强制修改 rememberMe 标记，让用户偏好保留
  })

  // 暴露必要的操作方法，便于在页面中调用（例如 Profile.vue 中更新头像等）
  return { token, userInfo, setUser, setToken, logout }
})
