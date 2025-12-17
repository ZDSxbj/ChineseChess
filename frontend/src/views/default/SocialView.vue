<script lang="ts" setup>
import type { ChatMessage } from '@/api/user/chat'
import type { FriendItem } from '@/api/user/social'
import type { UserInfo } from '@/api/user/user'
import { inject, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getMessages, markRead, sendMessage } from '@/api/user/chat'
import { getUserInfo } from '@/api/user/get_info'
import { acceptFriendRequest, checkFriendRequest, deleteFriend, deleteFriendRequest, getFriendChallenges, getFriendRequests, getFriends } from '@/api/user/social'
import FriendCard from '@/components/FriendCard.vue'
import FriendRequestCard from '@/components/FriendRequestCard.vue'
import { showMsg } from '@/components/MessageBox'
import { persistChallengeModals, restoreChallengeModals } from '@/store/challengeModalStore'
import { useSocialStore } from '@/store/useSocialStore'
// channel handling is centralized in social store
import { useUserStore } from '@/store/useStore'
import channel from '@/utils/channel'

const friends = ref<FriendItem[]>([])
const friendRequests = ref<any[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// Search state
const isSearchExpanded = ref(false)
const searchQuery = ref('')
const searchResult = ref<UserInfo | null>(null)
const searchStatus = ref<'idle' | 'searching' | 'found' | 'not-found' | 'self' | 'error'>('idle')

const selectedFriendId = ref<number | null>(null)
const messages = ref<ChatMessage[]>([])
const inputMsg = ref('')
const userStore = useUserStore()
const socialStore = useSocialStore()
const ws = inject('ws') as any
let _social_onMsg: ((data: any) => void) | null = null
const _base = import.meta.env.BASE_URL || '/'
const base = _base.endsWith('/') ? _base : `${_base}/`
const defaultAvatar = `${base}images/default_avatar.png`

// 好友挑战：发送方等待窗口
const waitingModal = ref<{ visible: boolean, challengeId?: number, receiverId?: number }>({ visible: false })
// 好友挑战：接收方响应窗口
const responseModal = ref<{ visible: boolean, challengeId?: number, senderId?: number, senderName?: string, roomId?: number }>({ visible: false })
// 模态状态持久化改由 store 提供函数
// 记录每个发送方的最近一次挑战元信息
const incomingChallengeMap = ref<Map<number, { challengeId: number, roomId: number, senderName?: string }>>(new Map())

async function loadMessagesForSelected() {
  const sid = selectedFriendId.value
  if (!sid)
    return
  const friend = friends.value.find(f => f.id === sid)
  if (!friend || !friend.relationId)
    return
  try {
    const resp: any = await getMessages(friend.relationId)
    // 支持后端直接返回数组或 { messages: [] }
    const arr = Array.isArray(resp) ? resp : (resp.messages || [])
    messages.value = arr
    // 标记为已读
    await markRead(friend.relationId)
    // 将对应好友的未读数清零，并从全局总数中扣除
    const prev = friend.unreadCount || 0
    if (prev > 0) {
      socialStore.decrease(prev)
    }
    friend.unreadCount = 0
  }
  catch (e) {
    console.error('加载聊天记录失败', e)
  }
  await nextTick()
  scrollToBottom()
}

// 当 messages 变化时，自动滚动到底部（避免在某些情况下错过）
watch(messages, () => {
  scrollToBottom()
})

const chatScrollRef = ref<HTMLDivElement | null>(null)
function scrollToBottom() {
  nextTick(() => {
    if (chatScrollRef.value)
      chatScrollRef.value.scrollTop = chatScrollRef.value.scrollHeight
  })
}

async function sendCurrentMessage() {
  const sid = selectedFriendId.value
  if (!sid)
    return
  const friend = friends.value.find(f => f.id === sid)
  if (!friend || !friend.relationId)
    return
  const text = inputMsg.value.trim()
  if (!text)
    return
  try {
    const resp: any = await sendMessage(friend.relationId, text)
    // 追加到本地消息
    messages.value.push({
      id: resp.id || Date.now(),
      friendId: friend.relationId,
      senderId: (userStore.userInfo?.id) || (resp.senderId || 0),
      receiverId: (resp.receiverId) || friend.id,
      isRead: true,
      content: text,
      createdAt: resp.createdAt || new Date().toISOString(),
    })
    inputMsg.value = ''
    scrollToBottom()
  }
  catch (e) {
    console.error('发送失败', e)
    // eslint-disable-next-line no-alert
    alert('发送失败')
  }
}

// context menu state
const contextMenu = ref<{ visible: boolean, x: number, y: number, id: number | null }>({ visible: false, x: 0, y: 0, id: null })
const confirmDeleteId = ref<number | null>(null)

function toggleSearch() {
  isSearchExpanded.value = !isSearchExpanded.value
  if (!isSearchExpanded.value) {
    clearSearch()
  }
}

function clearSearch() {
  searchQuery.value = ''
  searchResult.value = null
  searchStatus.value = 'idle'
}

async function performSearch() {
  const name = searchQuery.value.trim()
  if (!name) return

  // Reset selection and active chat when searching
  selectedFriendId.value = null
  // eslint-disable-next-line style/max-statements-per-line
  try { socialStore.setActiveRelationId(null) } catch { /* ignore */ }

  if (name === userStore.userInfo?.name) {
    searchStatus.value = 'self'
    searchResult.value = null
    return
  }

  searchStatus.value = 'searching'
  searchResult.value = null

  try {
    const resp = await getUserInfo({ name })
    searchResult.value = resp as unknown as UserInfo
    // 检查当前用户是否已向该用户发送好友申请
    try {
      const chk: any = await checkFriendRequest((searchResult.value as any).id)
      ;(searchResult as any).pending = chk.exists
    } catch {}
    searchStatus.value = 'found'
  } catch {
    searchStatus.value = 'not-found'
  }
}
// Add-friend modal state
const addModalVisible = ref(false)
const addModalContent = ref('')
const addModalSubmitting = ref(false)

async function tryAddFriend() {
  if (!searchResult.value) return
  try {
    const chk: any = await checkFriendRequest((searchResult.value as any).id)
    if (chk && chk.exists) {
      ;(searchResult as any).pending = true
      showMsg('好友申请已发送，请等待对方回应')
      return
    }
  } catch (err: any) {
    const status = err?.response?.status
    if (status === 404) {
      showMsg('检查申请状态的接口未找到，请重启后端或稍后重试')
      return
    }
    showMsg('检查好友申请状态失败，请稍后重试')
    return
  }
  // 不存在相同申请，打开模态
  openAddFriendModal()
}

function openAddFriendModal() {
  if (!searchResult.value) return
  addModalContent.value = ''
  addModalVisible.value = true
  // 禁止背景滚动
  // eslint-disable-next-line style/max-statements-per-line
  try { document.body.style.overflow = 'hidden' } catch {}
  // focus will be handled via ref in template (autofocus not reliable in some browsers)
}

function closeAddFriendModal() {
  addModalVisible.value = false
  addModalSubmitting.value = false
  // eslint-disable-next-line style/max-statements-per-line
  try { document.body.style.overflow = '' } catch {}
}

async function sendAddFriend() {
  if (!ws) {
    // eslint-disable-next-line no-alert
    alert('未连接到 WebSocket，无法发送')
    return
  }
  if (!searchResult.value) return
  // 在发送前向后端检查是否已有相同的好友申请（避免重复发送）
  try {
    const chk: any = await checkFriendRequest((searchResult.value as any).id)
    if (chk && chk.exists) {
      ;(searchResult as any).pending = true
      // eslint-disable-next-line no-alert
      alert('好友申请已发送，请等待对方回应')
      closeAddFriendModal()
      return
    }
  } catch (err: any) {
    // 如果后端检查接口不可用（404）或其他错误，给出明确提示并阻止发送，避免重复创建记录
    const status = err?.response?.status
    if (status === 404) {
      // eslint-disable-next-line no-alert
      alert('检查申请状态的接口未找到，请重启后端或稍后重试')
      return
    }
    // 其他错误则提示并阻止发送
    // eslint-disable-next-line no-alert
    alert('检查好友申请状态失败，请稍后重试')
    return
  }
  const content = (addModalContent.value || '').trim() || '你好，我们加个好友吧'
  addModalSubmitting.value = true
  try {
    ws.sendFriendRequest(searchResult.value.id, content)
    ;(searchResult as any).pending = true
    closeAddFriendModal()
    // 简短提示
    // eslint-disable-next-line no-alert
    alert('好友申请已发送')
  }
  catch (e) {
    console.error('发送好友申请失败', e)
    // eslint-disable-next-line no-alert
    alert('发送失败，请重试')
    addModalSubmitting.value = false
  }
}

let pollTimer: number | undefined

async function load(isPolling = false) {
  if (!isPolling)
    loading.value = true
  try {
    const resp = await getFriends()
    // 比较新旧好友列表，仅在数据发生变化时才更新，避免不必要的 UI 刷新
    const newFriends = resp.friends || []
    const oldFriends = friends.value || []

    function isSameFriend(a: typeof newFriends[0], b: typeof newFriends[0]) {
      if (!a || !b)
        return false
      return a.id === b.id && a.name === b.name && a.avatar === b.avatar && a.online === b.online && a.gender === b.gender && a.exp === b.exp && a.totalGames === b.totalGames && Number(a.winRate).toFixed(4) === Number(b.winRate).toFixed(4)
    }

    function friendsEqual(oldList: typeof oldFriends, newList: typeof newFriends) {
      if (oldList.length !== newList.length)
        return false
      const oldMap = new Map<number, typeof oldList[0]>()
      for (const ofr of oldList) {
        oldMap.set(ofr.id, ofr)
      }
      for (const nfr of newList) {
        const ofr = oldMap.get(nfr.id)
        if (!ofr)
          return false
        if (!isSameFriend(ofr, nfr))
          return false
      }
      return true
    }

    if (!friendsEqual(oldFriends, newFriends)) {
      friends.value = newFriends
      // 计算并同步全局未读总数
      try {
        const sum = (newFriends || []).reduce((s: number, it: any) => s + (Number(it.unreadCount) || 0), 0)
        // also include pending friend requests count so totalUnread = chat unread + friend requests
        let reqCount = 0
        try {
          const rr: any = await getFriendRequests()
          reqCount = (rr.requests || []).length
        } catch {}
        socialStore.setTotal(sum + reqCount)
      }
      catch (e) {
        // ignore
      }
    }
    // 无论好友列表是否变化，都初始化挑战标签：拉取我收到的挑战请求（receiver=me）
    try {
      const chResp: any = await getFriendChallenges()
      const challenges = chResp.challenges || []
      const set = new Set<number>()
      // 重置并填充 incomingChallengeMap
      incomingChallengeMap.value.clear()
      challenges.forEach((c: any) => {
        // 兼容后端字段命名（Go 默认导出字段为 PascalCase）
        const challengeId = c.id ?? c.ID
        const senderId = c.senderId ?? c.SenderID
        const roomId = (c as any).roomId ?? c.RoomID ?? 0
        if (senderId) {
          set.add(Number(senderId))
          incomingChallengeMap.value.set(Number(senderId), { challengeId: Number(challengeId), roomId: Number(roomId) })
        }
      })
      friends.value.forEach((f) => {
        // 若当前就在该好友的活跃聊天或响应弹窗已打开，则不显示标签（避免轮询覆盖）
        const isActiveChat = selectedFriendId.value === f.id
        const isResponding = responseModal.value.visible && responseModal.value.senderId === f.id
        ;(f as any).challengePending = (!isActiveChat && !isResponding) && set.has(f.id)
      })
    } catch {}
    error.value = null
  }
  catch (e) {
    error.value = (e as any)?.toString() || '获取好友失败'
  }
  finally {
    if (!isPolling)
      loading.value = false
  }
}

async function acceptRequest(requestId: number) {
  try {
    const resp: any = await acceptFriendRequest(requestId)
    // 减少未读
    try { socialStore.decrease(1) } catch {}
    // 刷新好友列表
    await load()
    // 后端已保存并推送招呼消息，前端无需重复发送
    // 移除本地请求列表项
    friendRequests.value = friendRequests.value.filter((r: any) => r.id !== requestId)
  } catch (e) {
    // eslint-disable-next-line no-alert
    alert('同意好友申请失败')
  }
}

async function rejectRequest(requestId: number) {
  try {
    await deleteFriendRequest(requestId)
    try { socialStore.decrease(1) } catch {}
    friendRequests.value = friendRequests.value.filter((r: any) => r.id !== requestId)
  } catch (e) {
    // eslint-disable-next-line no-alert
    alert('拒绝好友申请失败')
  }
}

function openContextMenu(payload: { id: number, x: number, y: number }) {
  contextMenu.value = { visible: true, x: payload.x, y: payload.y, id: payload.id }
}

function closeContextMenu() {
  contextMenu.value.visible = false
  contextMenu.value.id = null
}

function onSelectFriend(id: number) {
  selectedFriendId.value = id
  const friend = friends.value.find(f => f.id === id)
  // 通知 store 当前正在查看的 relationId（用于避免将当前会话的新消息计为未读）
  socialStore.setActiveRelationId(friend?.relationId || null)
  // 加载该好友的历史消息并标记为已读
  loadMessagesForSelected()
  // 若有挑战挂起，清除标签并弹出响应窗口
  if (friend && (friend as any).challengePending) {
    (friend as any).challengePending = false
    const meta = incomingChallengeMap.value.get(friend.id)
    if (meta) {
      responseModal.value = { visible: true, challengeId: meta.challengeId, senderId: friend.id, senderName: friend.name, roomId: meta.roomId }
    }
  }
}

function requestDelete(id: number) {
  confirmDeleteId.value = id
  closeContextMenu()
}

async function confirmDelete() {
  const id = confirmDeleteId.value
  if (!id)
    return
  try {
    await deleteFriend(id)
    friends.value = friends.value.filter(f => f.id !== id)
    if (selectedFriendId.value === id)
      selectedFriendId.value = null
      // 清除 store 中的 activeRelationId，避免离开会话后仍被认为活跃
    socialStore.setActiveRelationId(null)
    confirmDeleteId.value = null
  }
  catch {
    // 使用浏览器 alert 简单提示
    // eslint-disable-next-line no-alert
    alert('删除失败')
  }
}

// 点击对弈按钮
function onClickChallenge() {
  const sid = selectedFriendId.value
  if (!sid || !ws) return
  const f = friends.value.find(fr => fr.id === sid)
  if (!f || !f.relationId) return
  if (!f.online) {
    showMsg('对方不在线，需等待对方在线才可对战')
    return
  }
  // 发起挑战，后端会创建房间并推给对方
  try {
    ws.sendFriendChallengeInvite(f.id, f.relationId)
    // 打开等待窗口（challengeId 发送后由取消时需要，先缓存在收到服务端回执/事件前用不到，这里只显示等待）
    waitingModal.value = { visible: true }
    persistChallengeModals(waitingModal.value, responseModal.value)
  } catch {}
}

function onCancelWaiting() {
  if (!waitingModal.value.visible) return
  if (waitingModal.value.challengeId && waitingModal.value.receiverId && ws) {
    ws.sendFriendChallengeCancel(waitingModal.value.challengeId, waitingModal.value.receiverId)
  }
  waitingModal.value = { visible: false }
  persistChallengeModals(waitingModal.value, responseModal.value)
}

function onAcceptChallenge() {
  const meta = responseModal.value
  if (!meta.visible || !ws) return
  if (!meta.challengeId || !meta.senderId || !meta.roomId) {
    responseModal.value.visible = false
    return
  }
  // 通知后端接受，并加入房间
  ws.sendFriendChallengeAccept(meta.challengeId, meta.senderId, meta.roomId)
  // 客户端直接加入房间，确保及时进入对战
  ws.join(meta.roomId)
  responseModal.value.visible = false
  persistChallengeModals(waitingModal.value, responseModal.value)
}

function onRejectChallenge() {
  const meta = responseModal.value
  if (!meta.visible || !ws) return
  if (!meta.challengeId || !meta.senderId) {
    responseModal.value.visible = false
    return
  }
  ws.sendFriendChallengeReject(meta.challengeId, meta.senderId)
  responseModal.value.visible = false
  persistChallengeModals(waitingModal.value, responseModal.value)
}

onMounted(() => {
  // 刷新恢复模态状态
  const restored = restoreChallengeModals()
  if (restored.waiting) waitingModal.value = restored.waiting
  // 为了保持一致性：重新进入社交界面时不直接弹出响应窗口，
  // 而是仅在点击对应好友卡片后再弹出。
  // 因此当恢复到一个可见的响应模态时，将其转换为好友列表上的“挑战”标签。
  if (restored.response) {
    const r = restored.response
    // 始终关闭恢复的弹窗可见性
    responseModal.value = { visible: false }
    // 如果能拿到 senderId，则标记该好友有挂起挑战
    if (typeof r.senderId === 'number' && r.senderId > 0) {
      incomingChallengeMap.value.set(r.senderId, {
        challengeId: r.challengeId || 0,
        roomId: r.roomId || 0,
        senderName: r.senderName,
      })
      const f = friends.value.find(fr => fr.id === r.senderId)
      if (f) {
        ;(f as any).challengePending = true
      }
    }
  }
  load()
  // 轮询作为后备方案，建议服务端通过 WebSocket 推送在线状态更新以减少轮询
  pollTimer = window.setInterval(() => {
    load(true)
  }, 8000)

  // 初始化 social store（会订阅 channel 并计算初始未读）
  try {
    socialStore.init()
  }
  catch {
    // ignore
  }

  // 加载当前的好友申请列表
  async function loadFriendRequests() {
    try {
      const resp: any = await getFriendRequests()
      friendRequests.value = resp.requests || []
    } catch (e) {
      // ignore
    }
  }
  loadFriendRequests()

  // 订阅 websocket 的好友申请推送，及时显示新请求
  channel.on('NET:FRIEND:REQUEST', (data: any) => {
    try {
      // payload: { requestId, senderId, senderName, content, createdAt }
      const it = {
        id: data.requestId,
        senderId: data.senderId,
        senderName: data.senderName,
        senderAvatar: undefined,
        content: data.content,
        createdAt: data.createdAt,
      }
      // 将新申请放到最前面
      friendRequests.value.unshift(it)
    } catch (e) {}
  })


  // 好友对弈邀请
  channel.on('NET:FRIEND:CHALLENGE:INVITE', (payload: any) => {
    try {
      incomingChallengeMap.value.set(payload.senderId, { challengeId: payload.challengeId, roomId: payload.roomId, senderName: payload.senderName })
      const f = friends.value.find(fr => fr.id === payload.senderId)
      if (f) {
        if (selectedFriendId.value === f.id) {
          ;(f as any).challengePending = false
          responseModal.value = { visible: true, challengeId: payload.challengeId, senderId: payload.senderId, senderName: payload.senderName, roomId: payload.roomId }
          persistChallengeModals(waitingModal.value, responseModal.value)
        } else {
          ;(f as any).challengePending = true
        }
      }
    } catch {}
  })

  // 对方撤销挑战
  channel.on('NET:FRIEND:CHALLENGE:CANCEL', (payload: any) => {
    try {
      const senderId = Number(payload.senderId)
      incomingChallengeMap.value.delete(senderId)
      const f = friends.value.find(fr => fr.id === senderId)
      if (f) (f as any).challengePending = false
      if (responseModal.value.visible && responseModal.value.senderId === senderId) {
        responseModal.value.visible = false
        persistChallengeModals(waitingModal.value, responseModal.value)
      }
    } catch {}
  })

  // 发送方收到同意
  channel.on('NET:FRIEND:CHALLENGE:ACCEPT', () => {
    waitingModal.value.visible = false
    persistChallengeModals(waitingModal.value, responseModal.value)
  })

  // 发送方收到拒绝
  channel.on('NET:FRIEND:CHALLENGE:REJECT', () => {
    waitingModal.value.visible = false
    persistChallengeModals(waitingModal.value, responseModal.value)
  })

  // 发送方收到创建回执：记录 challengeId 以支持撤销
  channel.on('NET:FRIEND:CHALLENGE:CREATED', (payload: any) => {
    try {
      waitingModal.value = { visible: true, challengeId: payload.challengeId, receiverId: payload.receiverId }
      persistChallengeModals(waitingModal.value, responseModal.value)
    } catch {}
  })

  // 注册消息回调，用于将消息追加到当前打开的会话或更新对应好友的 unreadCount
  _social_onMsg = (data: any) => {
    const { relationId, senderId, sender, content, messageId, createdAt } = data || {}
    if (!content)
      return
    let f = null as any
    if (relationId) {
      f = friends.value.find((x: any) => x.relationId === relationId)
    }
    if (!f && sender) {
      f = friends.value.find((x: any) => x.name === sender || String(x.id) === sender)
    }
    if (f) {
      const msg = {
        id: messageId || Date.now(),
        friendId: f.relationId,
        senderId: senderId || 0,
        receiverId: f.id,
        isRead: selectedFriendId.value === f.id,
        content,
        createdAt: createdAt ? new Date(createdAt * 1000).toISOString() : new Date().toISOString(),
      }
      if (selectedFriendId.value === f.id) {
        messages.value.push(msg)
        // 标记后端已读（store 已在 init 的 channel handler 内处理，但这里也兜底）
        if (f.relationId)
          markRead(f.relationId)
        scrollToBottom()
      }
      else {
        // 仅更新该好友的本地 unreadCount（全局计数由 store 维护）
        f.unreadCount = (f.unreadCount || 0) + 1
      }
    }
  }

  if (_social_onMsg)
    socialStore.onMessage(_social_onMsg)
})

onBeforeUnmount(() => {
  if (pollTimer)
    window.clearInterval(pollTimer)
  // 移除消息回调
  if (_social_onMsg)
    socialStore.offMessage(_social_onMsg)
  // 离开社交页面时，取消任何活跃会话，防止空间外仍被认为处于活跃状态
  // eslint-disable-next-line style/max-statements-per-line
  try { socialStore.setActiveRelationId(null) } catch { /* ignore */ }
})
</script>

<template>
  <div class="h-full flex bg-[#fdf6e3] animate-fade-in">
    <!-- 左侧好友列表 -->
    <aside class="w-96 border-r border-amber-200 bg-amber-50/50 flex flex-col shadow-sm z-10">
      <div class="p-5 border-b border-amber-200 bg-amber-100/80 backdrop-blur-sm sticky top-0 z-10 flex items-center justify-between">
        <h2 class="text-xl font-black text-amber-900 tracking-wide">好友列表</h2>
        <div class="flex items-center gap-2">
          <div v-if="isSearchExpanded" class="flex items-center gap-2 animate-fade-in-right">
            <input
              v-model="searchQuery"
              @keyup.enter="performSearch"
              type="text"
              placeholder="输入用户名搜索"
              class="px-3 py-1 text-sm border border-amber-300 rounded-full focus:outline-none focus:border-amber-600 focus:ring-1 focus:ring-amber-600/20 transition-all w-40 bg-white/80 text-amber-900 placeholder-amber-900/40"
            />
            <button @click="performSearch" class="bg-transparent border-none outline-none p-1 text-amber-700 hover:text-amber-900 transition-colors cursor-pointer flex items-center justify-center" title="搜索">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
            </button>
            <button @click="toggleSearch" class="bg-transparent border-none outline-none p-1 text-amber-700 hover:text-red-600 transition-colors cursor-pointer flex items-center justify-center" title="关闭">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
          </div>
          <button v-else @click="toggleSearch" class="bg-transparent border-none outline-none p-1 text-amber-700 hover:text-amber-900 transition-colors cursor-pointer flex items-center justify-center" title="搜索好友">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto custom-scrollbar p-2">
        <!-- Search Results -->
        <template v-if="isSearchExpanded && searchStatus !== 'idle'">
          <div v-if="searchStatus === 'searching'" class="p-8 text-center text-amber-800/60 animate-pulse font-medium">搜索中...</div>
          <div v-else-if="searchStatus === 'not-found'" class="p-8 text-center text-amber-800/60 font-medium">未找到该用户</div>
          <div v-else-if="searchStatus === 'self'" class="p-8 text-center text-amber-800/60 font-medium">不能搜索自己</div>
          <div v-else-if="searchStatus === 'found' && searchResult" class="divide-y divide-amber-200/50">
            <!-- Check if friend -->
            <template v-if="friends.some(f => f.id === searchResult!.id)">
              <FriendCard
                :friend="friends.find(f => f.id === searchResult!.id)!"
                :selected="selectedFriendId === searchResult!.id"
                @select="onSelectFriend"
                @context="openContextMenu"
              />
            </template>
            <template v-else>
              <!-- Non-friend card -->
              <div class="flex items-center gap-4 p-4 bg-white/60 border-b border-amber-200/50 rounded-xl mb-2">
                <img
                  :src="searchResult.avatar || defaultAvatar"
                  alt="avatar"
                  class="w-14 h-14 rounded-full object-cover border-2 border-amber-200 shadow-sm"
                />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <h3 class="text-base font-bold text-amber-900 truncate">{{ searchResult.name }}</h3>
                      <span v-if="searchResult.gender === '男'" class="text-blue-600 text-lg font-bold" title="男">♂</span>
                      <span v-else-if="searchResult.gender === '女'" class="text-pink-600 text-lg font-bold" title="女">♀</span>
                      <span v-else-if="searchResult.gender === '其他'" class="text-gray-500 text-lg font-bold" title="其他">？</span>
                    </div>
                    <!-- Add Button -->
                    <button :disabled="(searchResult as any).pending" @click="tryAddFriend" class="px-4 py-1 text-xs font-bold text-white bg-amber-600 rounded-full hover:bg-amber-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-sm" :title="(searchResult as any).pending ? '请等待对方响应' : ''" aria-haspopup="dialog">
                      {{ (searchResult as any).pending ? '已申请' : '添加' }}
                    </button>
                  </div>

                  <div class="flex items-center gap-3 mt-1.5 text-sm text-amber-800/70 font-medium">
                    <span>场次: {{ searchResult.totalGames }}</span>
                    <span>胜率: {{ (searchResult.winRate || 0).toFixed(2) }}%</span>
                  </div>
                  <div class="text-xs text-amber-800/50 mt-0.5 font-medium">经验: {{ searchResult.exp }}</div>
                </div>
              </div>
            </template>
          </div>
        </template>

        <!-- Friend List -->
        <template v-else>
          <div v-if="loading" class="p-8 text-center text-amber-800/60 animate-pulse font-medium">加载中...</div>
          <div v-if="error" class="p-8 text-center text-red-600 bg-red-50 m-4 rounded-lg border border-red-100">{{ error }}</div>
          <div v-if="!loading && friends.length === 0" class="p-12 text-center text-amber-800/40 flex flex-col items-center gap-3">
            <div class="i-carbon-user-multiple text-4xl opacity-50"></div>
            <span class="font-medium">暂无好友</span>
          </div>
          <div class="space-y-1">
            <!-- Friend Requests -->
            <template v-if="friendRequests.length > 0">
              <div class="px-4 pt-3 pb-1 text-sm font-bold text-amber-800/60">好友申请</div>
              <div class="space-y-1">
                <FriendRequestCard
                  v-for="rq in friendRequests"
                  :key="rq.id"
                  :request="rq"
                  @accept="acceptRequest"
                  @reject="rejectRequest"
                />
              </div>
            </template>
            <FriendCard
              v-for="f in friends"
              :key="f.id"
              :friend="f"
              :selected="selectedFriendId === f.id"
              @select="onSelectFriend"
              @context="openContextMenu"
            />
          </div>
        </template>
      </div>
    </aside>

    <!-- 右侧聊天面板 -->
    <section class="flex-1 flex flex-col bg-[#fffdf5] relative">
      <div v-if="!selectedFriendId" class="absolute inset-0 flex items-center justify-center text-amber-800/30 bg-amber-50/30">
        <div class="text-center">
          <div class="i-carbon-chat text-6xl mb-4 mx-auto opacity-50"></div>
          <p class="text-xl font-bold tracking-widest">选择一个好友开始聊天</p>
        </div>
      </div>

      <template v-else>
        <div class="h-16 border-b border-amber-100 flex items-center px-6 bg-[#fffdf5] backdrop-blur-sm shadow-sm z-10">
          <h3 class="text-lg font-black text-amber-900 flex items-center gap-2">
            {{ friends.find(f => f.id === selectedFriendId)?.name }}
            <span v-if="friends.find(f => f.id === selectedFriendId)?.online" class="w-2.5 h-2.5 bg-green-600 rounded-full shadow-sm border border-white" title="在线"></span>
            <span v-else class="w-2.5 h-2.5 bg-gray-400 rounded-full border border-white" title="离线"></span>
          </h3>
        </div>

        <div ref="chatScrollRef" class="chat-body bg-[#fffdf5] p-6 custom-scrollbar">
          <div v-if="messages.length === 0" class="flex justify-center py-8">
            <span class="bg-amber-50 text-amber-800/60 px-4 py-1 rounded-full text-xs font-medium border border-amber-100">暂无聊天记录</span>
          </div>

          <div v-for="(msg, idx) in messages" :key="msg.id || idx" class="mb-4">
            <div class="message-row items-start" :class="[ (msg.senderId === (userStore.userInfo?.id || 0)) ? 'self' : 'other' ]">
              <!-- 如果是对方消息放左侧，自己的消息放右侧 -->
              <template v-if="msg.senderId !== (userStore.userInfo?.id || 0)">
                <img
                  :src="(friends.find(f => f.relationId === msg.friendId)?.avatar) || defaultAvatar"
                  alt="avatar"
                  class="w-10 h-10 rounded-full object-cover border-2 border-amber-200 shadow-sm mr-3"
                  @error="(e) => { const t = e.target as HTMLImageElement; if (!t.src.endsWith('images/default_avatar.png')) t.src = defaultAvatar }"
                />
              </template>

              <div class="bubble shadow-sm" :class="[ (msg.senderId === (userStore.userInfo?.id || 0)) ? 'bubble--self' : 'bubble--other' ]">
                <div class="bubble-content font-medium">{{ msg.content }}</div>
                <div class="bubble-time opacity-70">{{ new Date(msg.createdAt).toLocaleString() }}</div>
              </div>

              <template v-if="msg.senderId === (userStore.userInfo?.id || 0)">
                <img
                  :src="userStore.userInfo?.avatar || defaultAvatar"
                  alt="my-avatar"
                  class="w-10 h-10 rounded-full object-cover border-2 border-amber-200 shadow-sm ml-3"
                  @error="(e) => { const t = e.target as HTMLImageElement; if (!t.src.endsWith('images/default_avatar.png')) t.src = defaultAvatar }"
                />
              </template>
            </div>
          </div>
        </div>

        <div class="p-5 bg-[#fffdf5] border-t border-amber-100 backdrop-blur-sm">
          <div class="flex items-end gap-3 max-w-4xl mx-auto">
            <div class="flex-1 relative">
              <textarea
                v-model="inputMsg"
                @keyup.enter.exact.prevent="sendCurrentMessage"
                class="w-full border border-amber-200 rounded-xl p-3 pr-10 focus:ring-2 focus:ring-amber-500/20 focus:border-amber-500 outline-none resize-none bg-white transition-all text-amber-900 placeholder-amber-900/30 shadow-sm"
                placeholder="输入消息..."
                rows="3"
              ></textarea>
              <button class="absolute right-3 bottom-3 text-amber-400 hover:text-amber-600 p-1 rounded-full hover:bg-amber-50 transition-colors">
                <div class="i-carbon-face-add text-xl"></div>
              </button>
            </div>
            <div class="flex flex-col gap-2">
              <button @click="sendCurrentMessage" class="px-6 py-2.5 bg-gradient-to-r from-amber-600 to-amber-700 hover:from-amber-700 hover:to-amber-800 text-white rounded-lg shadow-md hover:shadow-lg transition-all font-bold flex items-center justify-center gap-2 active:scale-95 border border-amber-800/20">
                <div class="i-carbon-send-alt"></div>
                发送
              </button>
              <button @click="onClickChallenge" class="px-6 py-2.5 bg-gradient-to-r from-emerald-600 to-emerald-700 hover:from-emerald-700 hover:to-emerald-800 text-white rounded-lg shadow-md hover:shadow-lg transition-all font-bold flex items-center justify-center gap-2 active:scale-95 border border-emerald-800/20">
                <div class="i-carbon-game-console"></div>
                对弈
              </button>
            </div>
          </div>
        </div>
      </template>
    </section>

    <!-- 右键菜单遮罩 -->
    <div v-if="contextMenu.visible" class="fixed inset-0 z-40" @click="closeContextMenu"></div>

    <!-- 右键菜单 -->
    <div
      v-if="contextMenu.visible"
      :style="{ position: 'fixed', left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
      class="z-50 bg-white border border-amber-200 rounded-lg shadow-xl py-1 min-w-[120px] animate-fade-in"
    >
      <button class="w-full text-left px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition-colors flex items-center gap-2 font-medium" @click="requestDelete(contextMenu.id)">
        <div class="i-carbon-trash-can"></div>
        删除好友
      </button>
    </div>

    <!-- 确认删除弹窗 -->
    <div v-if="confirmDeleteId" class="fixed inset-0 flex items-center justify-center z-[60]">
      <div class="bg-black/60 backdrop-blur-sm absolute inset-0 transition-opacity" @click="confirmDeleteId = null"></div>
      <div class="bg-[#fdf6e3] p-6 rounded-xl shadow-2xl z-10 w-96 transform transition-all scale-100 border-2 border-amber-200">
        <h3 class="text-lg font-black text-amber-900 mb-2">确认删除</h3>
        <p class="text-amber-800/80 mb-6 font-medium">确定要删除该好友吗？此操作无法撤销。</p>
        <div class="flex justify-end gap-3">
          <button class="px-5 py-2 rounded-lg border-2 border-amber-200 text-amber-800 hover:bg-amber-100 font-bold transition-colors" @click="confirmDeleteId = null">取消</button>
          <button class="px-5 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white shadow-md hover:shadow-lg transition-all font-bold" @click="confirmDelete">确认删除</button>
        </div>
      </div>
    </div>

    <!-- 添加好友模态框 -->
    <div v-if="addModalVisible" class="fixed inset-0 z-[90] flex items-center justify-center">
      <!-- backdrop -->
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeAddFriendModal"></div>
      <div role="dialog" aria-modal="true" class="bg-[#fdf6e3] z-50 w-[420px] max-w-[90%] p-6 rounded-2xl shadow-2xl transform transition-all border-2 border-amber-200" @keydown.esc="closeAddFriendModal">
        <h3 class="text-lg font-black text-amber-900 mb-2">发送好友申请</h3>
        <p class="text-sm text-amber-800/70 mb-4 font-medium">给 <span class="font-bold text-amber-900">{{ searchResult?.name }}</span> 留言，让对方更容易接受你的申请。</p>
        <textarea v-model="addModalContent" rows="4" class="w-full border border-amber-300 rounded-lg p-3 resize-none focus:outline-none focus:ring-2 focus:ring-amber-500/30 bg-white text-amber-900 placeholder-amber-900/30" placeholder="说点什么吧（可选）"></textarea>
        <div class="flex items-center justify-end gap-3 mt-4">
          <button class="px-4 py-2 rounded-lg border-2 border-amber-200 text-amber-800 hover:bg-amber-100 font-bold" @click="closeAddFriendModal">取消</button>
          <button :disabled="addModalSubmitting" class="px-4 py-2 rounded-lg bg-amber-600 text-white hover:bg-amber-700 disabled:opacity-60 font-bold shadow-md" @click="sendAddFriend">发送</button>
        </div>
      </div>
    </div>
  </div>

  <!-- 发送方等待对方进入房间 -->
  <div v-if="waitingModal.visible" class="fixed inset-0 z-[95] flex items-center justify-center">
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-amber-50 p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <div class="w-16 h-16 mx-auto mb-4 text-amber-600 animate-spin">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
        </div>
        <h3 class="text-2xl font-black text-amber-900 mb-2">等待好友响应...</h3>
        <p class="text-amber-700 mb-6">已发送对弈邀请，请耐心等待</p>
        <button @click="onCancelWaiting" class="w-full bg-white border-2 border-amber-300 text-amber-800 py-3 rounded-xl font-bold hover:bg-amber-100 transition-colors">
          取消等待
        </button>
      </div>
    </div>
  </div>

  <!-- 接收方响应弹窗 -->
  <div v-if="responseModal.visible" class="fixed inset-0 z-[96] flex items-center justify-center">
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-amber-50 p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <h3 class="text-2xl font-black text-amber-900 mb-2">对弈邀请</h3>
        <p class="text-amber-700 mb-6"><span class="font-bold text-amber-900">{{ responseModal.senderName || '好友' }}</span> 邀请你进行一局对战</p>
        <div class="flex gap-3">
            <button @click="onRejectChallenge" class="flex-1 bg-white border-2 border-amber-300 text-amber-800 py-3 rounded-xl font-bold hover:bg-amber-100 transition-colors">
            拒绝
            </button>
            <button @click="onAcceptChallenge" class="flex-1 bg-gradient-to-r from-emerald-600 to-emerald-700 text-white py-3 rounded-xl font-bold hover:from-emerald-700 hover:to-emerald-800 transition-colors shadow-lg">
            接受挑战
            </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 确保 flex 子项能正确收缩以允许内部滚动 */
.chat-body {
  flex: 1 1 0%;
  min-height: 0;
  overflow-y: auto;
}

.message-row {
  display: flex;
}
.message-row.self {
  justify-content: flex-end;
}
.message-row.other {
  justify-content: flex-start;
}

.bubble {
  max-width: 70%;
  padding: 10px 12px;
  border-radius: 12px;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.03);
}
.bubble--self {
  background: linear-gradient(135deg, #d97706, #b45309);
  color: #fffbeb;
  border-bottom-right-radius: 4px;
  box-shadow: 0 2px 4px rgba(180, 83, 9, 0.2);
}
.bubble--other {
  background: #ffffff;
  color: #451a03;
  border: 1px solid #fde68a;
  border-bottom-left-radius: 4px;
}

.bubble-content {
  white-space: pre-wrap;
  word-break: break-word;
}
.bubble-time {
  font-size: 11px;
  color: currentColor;
  opacity: 0.7;
  margin-top: 6px;
  text-align: right;
}

/* 细微调整头像和气泡在小屏下的布局 */
@media (max-width: 640px) {
  .bubble { max-width: 85%; }
}

.custom-scrollbar::-webkit-scrollbar { width: 6px }
.custom-scrollbar::-webkit-scrollbar-track { background: #fffbeb }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #d97706; border-radius: 3px; opacity: 0.5; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #b45309 }
</style>
