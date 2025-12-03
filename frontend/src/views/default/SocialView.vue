<script lang="ts" setup>
import type { FriendItem } from '@/api/user/social'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { deleteFriend, getFriends } from '@/api/user/social'
import FriendCard from '@/components/FriendCard.vue'
import channel from '@/utils/channel'

const friends = ref<FriendItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const selectedFriendId = ref<number | null>(null)

// context menu state
const contextMenu = ref<{ visible: boolean; x: number; y: number; id: number | null }>({ visible: false, x: 0, y: 0, id: null })
const confirmDeleteId = ref<number | null>(null)

let pollTimer: number | undefined

async function load(isPolling = false) {
  if (!isPolling) loading.value = true
  try {
    const resp = await getFriends()
    // 比较新旧好友列表，仅在数据发生变化时才更新，避免不必要的 UI 刷新
    const newFriends = resp.friends || []
    const oldFriends = friends.value || []

    function isSameFriend(a: typeof newFriends[0], b: typeof newFriends[0]) {
      if (!a || !b) return false
      return a.id === b.id && a.name === b.name && a.avatar === b.avatar && a.online === b.online && a.gender === b.gender && a.exp === b.exp && a.totalGames === b.totalGames && Number(a.winRate).toFixed(4) === Number(b.winRate).toFixed(4)
    }

    function friendsEqual(oldList: typeof oldFriends, newList: typeof newFriends) {
      if (oldList.length !== newList.length) return false
      const oldMap = new Map<number, typeof oldList[0]>()
      for (const ofr of oldList) {
        oldMap.set(ofr.id, ofr)
      }
      for (const nfr of newList) {
        const ofr = oldMap.get(nfr.id)
        if (!ofr) return false
        if (!isSameFriend(ofr, nfr)) return false
      }
      return true
    }

    if (!friendsEqual(oldFriends, newFriends)) {
      friends.value = newFriends
    }
    error.value = null
  } catch (e) {
    error.value = (e as any)?.toString() || '获取好友失败'
  } finally {
    if (!isPolling) loading.value = false
  }
}

function openContextMenu(payload: { id: number; x: number; y: number }) {
  contextMenu.value = { visible: true, x: payload.x, y: payload.y, id: payload.id }
}

function closeContextMenu() {
  contextMenu.value.visible = false
  contextMenu.value.id = null
}

function onSelectFriend(id: number) {
  selectedFriendId.value = id
}

function requestDelete(id: number) {
  confirmDeleteId.value = id
  closeContextMenu()
}

async function confirmDelete() {
  const id = confirmDeleteId.value
  if (!id) return
  try {
    await deleteFriend(id)
    friends.value = friends.value.filter(f => f.id !== id)
    if (selectedFriendId.value === id) selectedFriendId.value = null
    confirmDeleteId.value = null
  } catch (e) {
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

  // // 订阅聊天消息事件，收到消息时可将对应好友标记为在线（作为实时性补充）
  // channel.on('NET:CHAT:MESSAGE', (data) => {
  //   const sender = data.sender
  //   if (!sender) return
  //   const f = friends.value.find(x => x.name === sender || String(x.id) === sender)
  //   if (f) f.online = true
  // })
})

onBeforeUnmount(() => {
  if (pollTimer) window.clearInterval(pollTimer)
  channel.off('NET:CHAT:MESSAGE')
})
</script>

<template>
  <div class="h-full flex bg-gray-50">
    <!-- 左侧好友列表 -->
    <aside class="w-96 border-r bg-white flex flex-col shadow-sm z-10">
      <div class="p-5 border-b bg-gray-50/50 backdrop-blur-sm sticky top-0 z-10">
        <h2 class="text-xl font-bold text-gray-800">好友列表</h2>
      </div>
      <div class="flex-1 overflow-y-auto custom-scrollbar">
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

        <div class="flex-1 bg-gray-50 p-6 overflow-y-auto custom-scrollbar">
          <!-- 聊天消息展示占位 -->
          <div class="flex justify-center py-8">
            <span class="bg-gray-200 text-gray-500 px-4 py-1 rounded-full text-xs">聊天记录加载完毕</span>
          </div>
        </div>

        <div class="p-5 bg-white border-t">
          <div class="flex items-end gap-3 max-w-4xl mx-auto">
            <div class="flex-1 relative">
              <textarea
                class="w-full border border-gray-300 rounded-xl p-3 pr-10 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 outline-none resize-none bg-gray-50 focus:bg-white transition-all"
                placeholder="输入消息..."
                rows="3"
              ></textarea>
              <button class="absolute right-3 bottom-3 text-gray-400 hover:text-gray-600 p-1 rounded-full hover:bg-gray-100 transition-colors">
                <div class="i-carbon-face-add text-xl"></div>
              </button>
            </div>
            <div class="flex flex-col gap-2">
              <button class="px-6 py-2.5 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-700 hover:to-blue-600 text-white rounded-lg shadow-md hover:shadow-lg transition-all font-medium flex items-center justify-center gap-2 active:scale-95">
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
      :style="{ position: 'fixed', left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
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
