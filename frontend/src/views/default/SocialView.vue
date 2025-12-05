<script lang="ts" setup>
import type { ChatMessage } from '@/api/user/chat'
import type { FriendItem } from '@/api/user/social'
import type { UserInfo } from '@/api/user/user'
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getMessages, markRead, sendMessage } from '@/api/user/chat'
import { deleteFriend, getFriends } from '@/api/user/social'
import { getUserInfo } from '@/api/user/get_info'
import FriendCard from '@/components/FriendCard.vue'
// channel handling is centralized in social store
import { useUserStore } from '@/store/useStore'
import { useSocialStore } from '@/store/useSocialStore'

const friends = ref<FriendItem[]>([])
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
let _social_onMsg: ((data: any) => void) | null = null
const _base = import.meta.env.BASE_URL || '/'
const base = _base.endsWith('/') ? _base : `${_base}/`
const defaultAvatar = `${base}images/default_avatar.png`

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
    searchStatus.value = 'found'
  } catch (e) {
    searchStatus.value = 'not-found'
  }
}function handleAddFriend() {
  // TODO: Implement add friend logic
  alert('添加好友功能尚未实现')
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
        socialStore.setTotal(sum)
      }
      catch (e) {
        // ignore
      }
    }
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

onMounted(() => {
  load()
  // 轮询作为后备方案，建议服务端通过 WebSocket 推送在线状态更新以减少轮询
  pollTimer = window.setInterval(() => {
    load(true)
  }, 8000)

  // 初始化 social store（会订阅 channel 并计算初始未读）
  try {
    socialStore.init()
  }
  catch (e) {
    // ignore
  }

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
  try { socialStore.setActiveRelationId(null) } catch { /* ignore */ }
})
</script>

<template>
  <div class="h-full flex bg-gray-50">
    <!-- 左侧好友列表 -->
    <aside class="w-96 border-r bg-white flex flex-col shadow-sm z-10">
      <div class="p-5 border-b bg-gray-50/50 backdrop-blur-sm sticky top-0 z-10 flex items-center justify-between">
        <h2 class="text-xl font-bold text-gray-800">好友列表</h2>
        <div class="flex items-center gap-2">
          <div v-if="isSearchExpanded" class="flex items-center gap-2 animate-fade-in-right">
            <input
              v-model="searchQuery"
              @keyup.enter="performSearch"
              type="text"
              placeholder="输入用户名搜索"
              class="px-3 py-1 text-sm border rounded-full focus:outline-none focus:border-blue-500 transition-all w-40"
            />
            <button @click="performSearch" class="bg-transparent border-none outline-none p-1 text-gray-500 hover:text-blue-500 transition-colors cursor-pointer flex items-center justify-center" title="搜索">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
            </button>
            <button @click="toggleSearch" class="bg-transparent border-none outline-none p-1 text-gray-500 hover:text-red-500 transition-colors cursor-pointer flex items-center justify-center" title="关闭">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
            </button>
          </div>
          <button v-else @click="toggleSearch" class="bg-transparent border-none outline-none p-1 text-gray-500 hover:text-blue-500 transition-colors cursor-pointer flex items-center justify-center" title="搜索好友">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto custom-scrollbar">
        <!-- Search Results -->
        <template v-if="isSearchExpanded && searchStatus !== 'idle'">
          <div v-if="searchStatus === 'searching'" class="p-8 text-center text-gray-500 animate-pulse">搜索中...</div>
          <div v-else-if="searchStatus === 'not-found'" class="p-8 text-center text-gray-500">未找到该用户</div>
          <div v-else-if="searchStatus === 'self'" class="p-8 text-center text-gray-500">不能搜索自己</div>
          <div v-else-if="searchStatus === 'found' && searchResult" class="divide-y divide-gray-50">
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
              <div class="flex items-center gap-4 p-4 bg-white border-b border-gray-100">
                <img
                  :src="searchResult.avatar || defaultAvatar"
                  alt="avatar"
                  class="w-14 h-14 rounded-full object-cover border border-gray-200 shadow-sm"
                />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2">
                      <h3 class="text-base font-semibold text-gray-800 truncate">{{ searchResult.name }}</h3>
                      <div v-if="searchResult.gender === '男'" class="i-carbon-gender-male text-blue-500 text-lg" title="男"></div>
                      <div v-else-if="searchResult.gender === '女'" class="i-carbon-gender-female text-pink-500 text-lg" title="女"></div>
                      <div v-else-if="searchResult.gender === '其他'" class="i-carbon-help text-gray-400 text-lg" title="其他"></div>
                    </div>
                    <!-- Add Button -->
                    <button @click="handleAddFriend" class="px-3 py-1 text-xs text-white bg-blue-500 rounded-full hover:bg-blue-600 transition-colors">
                      添加
                    </button>
                  </div>

                  <div class="flex items-center gap-3 mt-1.5 text-sm text-gray-500">
                    <span>场次: {{ searchResult.totalGames }}</span>
                    <span>胜率: {{ (searchResult.winRate * 100).toFixed(1) }}%</span>
                  </div>
                  <div class="text-xs text-gray-400 mt-0.5">经验: {{ searchResult.exp }}</div>
                </div>
              </div>
            </template>
          </div>
        </template>

        <!-- Friend List -->
        <template v-else>
          <div v-if="loading" class="p-8 text-center text-gray-500 animate-pulse">加载中...</div>
          <div v-if="error" class="p-8 text-center text-red-500 bg-red-50 m-4 rounded-lg">{{ error }}</div>
          <div v-if="!loading && friends.length === 0" class="p-12 text-center text-gray-400 flex flex-col items-center gap-3">
            <div class="i-carbon-user-multiple text-4xl opacity-50"></div>
            <span>暂无好友</span>
          </div>
          <div class="divide-y divide-gray-50">
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
    <section class="flex-1 flex flex-col bg-white relative">
      <div v-if="!selectedFriendId" class="absolute inset-0 flex items-center justify-center text-gray-300 bg-gray-50/50">
        <div class="text-center">
          <div class="i-carbon-chat text-6xl mb-4 mx-auto opacity-50"></div>
          <p class="text-xl font-medium">选择一个好友开始聊天</p>
        </div>
      </div>

      <template v-else>
        <div class="h-16 border-b flex items-center px-6 bg-white shadow-sm z-10">
          <h3 class="text-lg font-bold text-gray-800 flex items-center gap-2">
            {{ friends.find(f => f.id === selectedFriendId)?.name }}
            <span v-if="friends.find(f => f.id === selectedFriendId)?.online" class="w-2 h-2 bg-green-500 rounded-full shadow-green-200 shadow-lg"></span>
          </h3>
        </div>

        <div ref="chatScrollRef" class="chat-body bg-gray-50 p-6 custom-scrollbar">
          <div v-if="messages.length === 0" class="flex justify-center py-8">
            <span class="bg-gray-200 text-gray-500 px-4 py-1 rounded-full text-xs">暂无聊天记录</span>
          </div>

          <div v-for="(msg, idx) in messages" :key="msg.id || idx" class="mb-4">
            <div class="message-row items-start" :class="[ (msg.senderId === (userStore.userInfo?.id || 0)) ? 'self' : 'other' ]">
              <!-- 如果是对方消息放左侧，自己的消息放右侧 -->
              <template v-if="msg.senderId !== (userStore.userInfo?.id || 0)">
                <img
                  :src="(friends.find(f => f.relationId === msg.friendId)?.avatar) || defaultAvatar"
                  alt="avatar"
                  class="w-10 h-10 rounded-full object-cover border border-gray-200 shadow-sm mr-3"
                  @error="(e) => { const t = e.target as HTMLImageElement; if (!t.src.endsWith('images/default_avatar.png')) t.src = defaultAvatar }"
                />
              </template>

              <div class="bubble" :class="[ (msg.senderId === (userStore.userInfo?.id || 0)) ? 'bubble--self' : 'bubble--other' ]">
                <div class="bubble-content">{{ msg.content }}</div>
                <div class="bubble-time">{{ new Date(msg.createdAt).toLocaleString() }}</div>
              </div>

              <template v-if="msg.senderId === (userStore.userInfo?.id || 0)">
                <img
                  :src="userStore.userInfo?.avatar || defaultAvatar"
                  alt="my-avatar"
                  class="w-10 h-10 rounded-full object-cover border border-gray-200 shadow-sm ml-3"
                  @error="(e) => { const t = e.target as HTMLImageElement; if (!t.src.endsWith('images/default_avatar.png')) t.src = defaultAvatar }"
                />
              </template>
            </div>
          </div>
        </div>

        <div class="p-5 bg-white border-t">
          <div class="flex items-end gap-3 max-w-4xl mx-auto">
            <div class="flex-1 relative">
              <textarea
                v-model="inputMsg"
                @keyup.enter.exact.prevent="sendCurrentMessage"
                class="w-full border border-gray-300 rounded-xl p-3 pr-10 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none resize-none bg-gray-50 focus:bg-white transition-all"
                placeholder="输入消息..."
                rows="3"
              ></textarea>
              <button class="absolute right-3 bottom-3 text-gray-400 hover:text-gray-600 p-1 rounded-full hover:bg-gray-100 transition-colors">
                <div class="i-carbon-face-add text-xl"></div>
              </button>
            </div>
            <div class="flex flex-col gap-2">
              <button @click="sendCurrentMessage" class="px-6 py-2.5 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-700 hover:to-blue-600 text-white rounded-lg shadow-md hover:shadow-lg transition-all font-medium flex items-center justify-center gap-2 active:scale-95">
                <div class="i-carbon-send-alt"></div>
                发送
              </button>
              <button class="px-6 py-2.5 bg-gradient-to-r from-emerald-500 to-green-500 hover:from-emerald-600 hover:to-green-600 text-white rounded-lg shadow-md hover:shadow-lg transition-all font-medium flex items-center justify-center gap-2 active:scale-95">
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
      class="z-50 bg-white border border-gray-100 rounded-lg shadow-xl py-1 min-w-[120px] animate-fade-in"
    >
      <button class="w-full text-left px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition-colors flex items-center gap-2" @click="requestDelete(contextMenu.id)">
        <div class="i-carbon-trash-can"></div>
        删除好友
      </button>
    </div>

    <!-- 确认删除弹窗 -->
    <div v-if="confirmDeleteId" class="fixed inset-0 flex items-center justify-center z-[60]">
      <div class="bg-black/20 backdrop-blur-sm absolute inset-0 transition-opacity" @click="confirmDeleteId = null"></div>
      <div class="bg-white p-6 rounded-xl shadow-2xl z-10 w-96 transform transition-all scale-100">
        <h3 class="text-lg font-bold text-gray-800 mb-2">确认删除</h3>
        <p class="text-gray-600 mb-6">确定要删除该好友吗？此操作无法撤销。</p>
        <div class="flex justify-end gap-3">
          <button class="px-5 py-2 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium transition-colors" @click="confirmDeleteId = null">取消</button>
          <button class="px-5 py-2 rounded-lg bg-red-500 hover:bg-red-600 text-white shadow-md hover:shadow-lg transition-all font-medium" @click="confirmDelete">确认删除</button>
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
  background: linear-gradient(180deg,#007bff,#006ae6);
  color: white;
  border-bottom-right-radius: 4px;
}
.bubble--other {
  background: #f1f5f9;
  color: #111827;
  border-bottom-left-radius: 4px;
}

.bubble-content {
  white-space: pre-wrap;
  word-break: break-word;
}
.bubble-time {
  font-size: 11px;
  color: rgba(0,0,0,0.45);
  margin-top: 6px;
  text-align: right;
}

/* 细微调整头像和气泡在小屏下的布局 */
@media (max-width: 640px) {
  .bubble { max-width: 85%; }
}

.custom-scrollbar::-webkit-scrollbar { width: 6px }
.custom-scrollbar::-webkit-scrollbar-track { background: #f7fafc }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #cbd5e0; border-radius: 3px }
</style>
