import { defineStore } from 'pinia'
import { ref } from 'vue'
import { markRead } from '@/api/user/chat'
import { getFriendRequests, getFriends } from '@/api/user/social'
import channel from '@/utils/channel'

type MsgCallback = (data: any) => void

export const useSocialStore = defineStore('social', () => {
  const totalUnread = ref<number>(0)
  const activeRelationId = ref<number | null>(null)

  const listeners = new Set<MsgCallback>()
  let _inited = false

  function setTotal(n: number) {
    totalUnread.value = Math.max(0, Math.floor(n) || 0)
  }

  function increase(n = 1) {
    totalUnread.value = Math.max(0, totalUnread.value + Math.floor(n))
  }

  function decrease(n = 1) {
    totalUnread.value = Math.max(0, totalUnread.value - Math.floor(n))
  }

  function setActiveRelationId(id: number | null) {
    activeRelationId.value = id
  }

  function onMessage(cb: MsgCallback) {
    listeners.add(cb)
  }

  function offMessage(cb: MsgCallback) {
    listeners.delete(cb)
  }

  // 初始化：尝试从后端获取好友列表并计算未读总数，且订阅 channel 事件
  async function init() {
    if (_inited)
      return
    _inited = true
    try {
      const resp: any = await getFriends()
      const friends = resp.friends || []
      const sum = (friends || []).reduce((s: number, it: any) => s + (Number(it.unreadCount) || 0), 0)
      // also add pending friend requests
      let reqCount = 0
      try {
        const r: any = await getFriendRequests()
        reqCount = (r.requests || []).length
      } catch {}
      setTotal(sum + reqCount)
    }
    catch (e) {
      // 忽略错误（可能未登录或网络问题），稍后消息到来会补偿
    }

    // 注册 channel 监听
    channel.on('NET:CHAT:MESSAGE', (data: any) => {
      try {
        const { relationId } = data || {}
        // 仅社交会话（存在 relationId）参与未读统计；房间内聊天不计入
        if (typeof relationId === 'number' && relationId > 0) {
          // 如果当前正查看该会话，则直接标为已读（后端）且不增加总数
          if (activeRelationId.value === relationId) {
            try { markRead(relationId) } catch { /* ignore */ }
          } else {
            increase(1)
          }
        }
      }
      finally {
        // 通知所有注册的回调（用于让界面追加消息或更新单个好友的 unreadCount）
        listeners.forEach((cb) => { try { cb(data) } catch { /* ignore */ } })
      }
    })
    // 当收到好友申请推送时，增加未读计数
    channel.on('NET:FRIEND:REQUEST', (data: any) => {
      try {
        increase(1)
      } catch {}
    })
  }

  return { totalUnread, setTotal, increase, decrease, init, setActiveRelationId, onMessage, offMessage }
})
