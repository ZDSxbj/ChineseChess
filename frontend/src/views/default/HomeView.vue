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
  <div class="absolute inset-0 flex flex-col items-center justify-center p-6 text-center bg-[url('https://images.unsplash.com/photo-1528154109403-12502c3855a4?q=80&w=1000&auto=format&fit=crop')] bg-cover bg-center">
    <div class="absolute inset-0 bg-amber-900/80 backdrop-blur-sm"></div>

    <div class="relative z-10 text-amber-50 max-w-lg w-full space-y-8 animate-slide-up">
        <div class="mb-8">
            <div class="w-24 h-24 mx-auto bg-amber-100 text-amber-900 rounded-2xl flex items-center justify-center text-5xl font-black border-4 border-amber-600 shadow-2xl mb-4 transform rotate-3">
                象
            </div>
            <h1 class="text-4xl font-black tracking-[0.5em] mb-2">中国象棋</h1>
            <p class="text-amber-200 text-lg font-light">方寸之间，运筹帷幄</p>
        </div>

        <div class="bg-black/20 p-6 rounded-xl backdrop-blur-md border border-white/10 text-left mb-8 shadow-lg">
            <h3 class="text-xl font-bold mb-2 text-amber-300">简介</h3>
            <p class="text-sm leading-relaxed opacity-90 indent-8">
                中国象棋是起源于中国的一种棋戏，属于二人对抗性游戏的一种。它模拟了古代战争中两军对垒的场景，不仅趣味性强，而且能锻炼思维能力。在这里，你可以体验残局破解的快感，也可以与全球棋友实时对弈。
            </p>
        </div>

        <div class="grid grid-cols-1 gap-4 w-full px-4">
            <!-- Online Match Button -->
            <button
                @click="onlinePlay"
                class="group relative overflow-hidden bg-gradient-to-r from-amber-600 to-amber-700 text-white py-6 px-8 rounded-xl font-bold text-xl shadow-lg hover:scale-105 transition-all active:scale-95 flex items-center justify-center gap-3 w-full"
            >
                <span class="absolute inset-0 bg-white/20 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-500"></span>
                <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><path d="M10.5 14.5l-2.5 2.5"/><path d="M13 12l2.5-2.5"/></svg>
                随机匹配对战
            </button>

            <div class="grid grid-cols-2 gap-4">
                <button @click="singlePlay" class="bg-amber-100 text-amber-900 py-4 rounded-xl font-bold text-lg hover:bg-white transition-colors shadow-md flex items-center justify-center gap-2 hover:shadow-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
                    本地对战
                </button>
                <button @click="aiPlay" class="bg-amber-100 text-amber-900 py-4 rounded-xl font-bold text-lg hover:bg-white transition-colors shadow-md flex items-center justify-center gap-2 hover:shadow-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"></rect><rect x="9" y="9" width="6" height="6"></rect><line x1="9" y1="1" x2="9" y2="4"></line><line x1="15" y1="1" x2="15" y2="4"></line><line x1="9" y1="20" x2="9" y2="23"></line><line x1="15" y1="20" x2="15" y2="23"></line><line x1="20" y1="9" x2="23" y2="9"></line><line x1="20" y1="14" x2="23" y2="14"></line><line x1="1" y1="9" x2="4" y2="9"></line><line x1="1" y1="14" x2="4" y2="14"></line></svg>
                    人机对战
                </button>
            </div>
        </div>
    </div>

    <!-- Matching Modal -->
    <div v-if="matching" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div class="bg-amber-50 p-8 rounded-2xl shadow-2xl max-w-sm w-full text-center border-4 border-amber-200 animate-scale-in">
        <div class="w-16 h-16 mx-auto mb-4 text-amber-600 animate-spin">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
        </div>
        <h3 class="text-2xl font-black text-amber-900 mb-2">正在匹配对手...</h3>
        <p class="text-amber-700 mb-6">请稍候，正在为您寻找旗鼓相当的对手</p>
        <button @click="cancelMatch" class="w-full bg-white border-2 border-amber-300 text-amber-800 py-3 rounded-xl font-bold hover:bg-amber-100 transition-colors">
          取消匹配
        </button>
      </div>
    </div>
  </div>
</template>

