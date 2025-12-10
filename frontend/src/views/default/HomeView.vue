<script lang="ts" setup>
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import channel from '@/utils/channel'
import { showMsg } from '@/components/MessageBox'
import { useUserStore } from '@/store/useStore'

const router = useRouter()
const ws = inject('ws') as WebSocketService
const userStore = useUserStore()
const matching = ref(false)
const onMatchSuccess = () => {
  matching.value = false
}

function singlePlay() {
  channel.emit('MATCH:SUCCESS', null)
}

function aiPlay() {
  router.push('/game/ai-prepare')
}

function onlinePlay() {
  // 【问题1修复】校验登录状态
  if (!userStore.token || !userStore.userInfo?.id) {
    showMsg('请先登录后再进行在线匹配')
    return
  }
  
  if (matching.value)
    return
  matching.value = true
  ws?.match()
}

function cancelMatch() {
  // 【问题2修复】调用WebSocket取消匹配接口
  if (ws?.cancelMatch) {
    ws.cancelMatch()
  }
  matching.value = false
  showMsg('已取消匹配')
}

function handlePopState(_event: PopStateEvent) {
  window.history.pushState(null, '', window.location.href)
}

onMounted(() => {
  window.history.pushState(null, '', window.location.href)
  // 监听 popstate 事件，防止后退操作
  window.addEventListener('popstate', (event) => {
    handlePopState(event)
  })
  channel.on('MATCH:SUCCESS', onMatchSuccess)
})

onUnmounted(() => {
  window.removeEventListener('popstate', handlePopState)
  channel.off('MATCH:SUCCESS', onMatchSuccess)
})
</script>

<template>
  <main class="mx-a h-full w-9/10 bg-gray-4 p-1 sm:w-3/5">
    <h2 class="mx-a block w-fit">
      普普通通的首页
    </h2>
    <article class="text-xl line-height-9">
      <p>
        象棋与国际象棋及围棋并列世界三大棋类之一。象棋主要流行于全球华人、越南人及琉球人社区，是首届世界智力运动会的正式比赛项目之一。
      </p>
      <p>
        本网站是一个简易的象棋对战平台，支持本地对战、人机对战和在线对战。在线对战需要注册账号，或者以游客登录。
      </p>
      <p>更多功能待开发...也可能不会</p>
      <p>点击下方的按钮选择游戏模式。</p>
    </article>
    <div class="mt-4 flex flex-col gap-3 sm:flex-row">
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        :disabled="matching"
        @click="singlePlay"
      >
        本地对战
      </button>
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        :disabled="matching"
        @click="aiPlay"
      >
        人机对战
      </button>
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        :disabled="matching"
        @click="onlinePlay"
      >
        随机匹配
      </button>
    </div>
  </main>
  <div v-if="matching" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="w-80 rounded-lg bg-white p-6 shadow-lg text-center space-y-4">
      <div class="mx-auto h-10 w-10 animate-spin rounded-full border-4 border-gray-200 border-t-blue-500" />
      <div class="text-lg font-semibold">正在匹配...</div>
      <p class="text-sm text-gray-600">请稍等，系统正在为你寻找对手</p>
      <button class="w-full rounded bg-gray-700 px-4 py-2 text-white" @click="cancelMatch">取消匹配</button>
    </div>
  </div>
</template>

