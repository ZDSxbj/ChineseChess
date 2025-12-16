<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showMsg } from '@/components/MessageBox'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState, getGameState } from '@/store/gameStore'

const router = useRouter()
const background = ref<HTMLCanvasElement | null>(null)
const chesses = ref<HTMLCanvasElement | null>(null)

let chessBoard: ChessBoard | undefined
const hist = ref<any[]>([])
const index = ref(0)

function applySteps(n: number) {
  if (!chessBoard) {
    return
  }
  const saved = getGameState()
  if (!saved) {
    return
  }
  // 复位棋盘
  chessBoard.stop()
  chessBoard.start(saved.selfColor, false)
  // 逐步回放
  for (let i = 0; i < n; i++) {
    const step = hist.value[i]
    // 使用复盘专用的移动函数，仅改变棋盘状态并绘制，不触发其他逻辑
    chessBoard.replayMove(step.from, step.to)
  }
  chessBoard.disableInteraction()
}

function onExit() {
  clearGameState()
  router.push('/')
}

function onRestart() {
  index.value = 0
  applySteps(0)
}

function onPrev() {
  if (index.value <= 0) {
    return
  }
  index.value--
  applySteps(index.value)
}

function onNext() {
  const max = hist.value.length
  if (index.value >= max) {
    return
  }
  index.value++
  applySteps(index.value)
}
function handlePopState(_event: PopStateEvent) {
  globalThis.history.pushState(null, '', globalThis.location.href)
  showMsg('请通过应用内的导航按钮进行操作')
}

onMounted(() => {
  const saved = getGameState()
  if (!saved) {
    showMsg('未找到对局历史，无法进入复盘')
    router.push('/')
    return
  }

  const canvasBg = background.value as HTMLCanvasElement
  const canvasCh = chesses.value as HTMLCanvasElement
  if (!canvasBg || !canvasCh) {
    showMsg('画布初始化失败')
    router.push('/')
    return
  }

  chessBoard = new ChessBoard(canvasBg, canvasCh, 70)
  chessBoard.start(saved.selfColor, false)
  hist.value = saved.moveHistory || []
  index.value = hist.value.length // 默认到最后一步
  applySteps(index.value)
  globalThis.history.pushState(null, '', globalThis.location.href)
  // 监听 popstate 事件，防止后退操作
  globalThis.addEventListener('popstate', (event) => {
    handlePopState(event)
  })
})

onUnmounted(() => {
  chessBoard?.stop()
  globalThis.removeEventListener('popstate', handlePopState)
})
</script>

<template>
  <div class="h-full w-full bg-[#fdf6e3] flex flex-col items-center justify-center relative overflow-hidden">
    <!-- 背景装饰 -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute -top-[20%] -left-[10%] w-[70%] h-[70%] rounded-full bg-amber-200/20 blur-3xl"></div>
      <div class="absolute top-[40%] -right-[10%] w-[60%] h-[60%] rounded-full bg-orange-200/20 blur-3xl"></div>
    </div>

    <!-- 主内容 -->
    <div class="relative z-10 flex flex-col items-center justify-center w-full max-w-4xl p-4">
      <!-- 标题 -->
      <div class="mb-6 flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-amber-100 flex items-center justify-center text-amber-800 shadow-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h2 class="text-2xl font-black text-amber-900 tracking-wider">
          对局复盘
        </h2>
      </div>

      <!-- 棋盘容器 -->
      <div class="relative w-full max-w-[650px] aspect-[9/10] flex items-center justify-center mb-8">
        <!-- 棋盘背景装饰 -->
        <div class="absolute inset-4 bg-[#eecfa1] rounded shadow-2xl transform rotate-0 opacity-50 blur-sm"></div>

        <canvas
          ref="background"
          class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 shadow-2xl rounded-lg"
        />
        <canvas ref="chesses" class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 pointer-events-auto" />
      </div>

      <!-- 控制栏 -->
      <div class="bg-white/60 backdrop-blur-md border border-white/50 rounded-2xl p-4 shadow-lg w-full max-w-[600px]">
        <div class="flex flex-wrap justify-center gap-3 mb-3">
          <button
            class="group relative px-5 py-2 bg-gray-100 text-gray-700 rounded-xl font-bold shadow-sm hover:bg-gray-200 hover:shadow-md transition-all duration-200 flex items-center gap-2 border border-gray-200"
            @click="onExit"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            退出
          </button>

          <button
            class="group relative px-5 py-2 bg-blue-50 text-blue-700 rounded-xl font-bold shadow-sm hover:bg-blue-100 hover:shadow-md transition-all duration-200 flex items-center gap-2 border border-blue-100"
            @click="onRestart"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            重新开始
          </button>

          <button
            :disabled="index <= 0"
            class="group relative px-5 py-2 bg-amber-100 text-amber-900 rounded-xl font-bold shadow-sm hover:bg-amber-200 hover:shadow-md transition-all duration-200 flex items-center gap-2 border border-amber-200 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="onPrev"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            上一步
          </button>

          <button
            :disabled="index >= hist.length"
            class="group relative px-5 py-2 bg-emerald-50 text-emerald-700 rounded-xl font-bold shadow-sm hover:bg-emerald-100 hover:shadow-md transition-all duration-200 flex items-center gap-2 border border-emerald-100 disabled:opacity-50 disabled:cursor-not-allowed"
            @click="onNext"
          >
            下一步
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>

        <div class="text-center">
          <span class="text-xs font-bold text-amber-800/60 bg-amber-50 px-3 py-1 rounded-full border border-amber-100">
            第 {{ index }} 步 / 共 {{ hist.length }} 步
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
