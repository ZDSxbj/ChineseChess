<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showMsg } from '@/components/MessageBox'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState, getGameState } from '@/store/gameStore'
import channel from '@/utils/channel'

const router = useRouter()
const background = ref<HTMLCanvasElement | null>(null)
const chesses = ref<HTMLCanvasElement | null>(null)

let chessBoard: ChessBoard | undefined
const hist = ref<any[]>([])
const index = ref(0)

function applySteps(n: number) {
  if (!chessBoard) return
  const saved = getGameState()
  if (!saved) return
  // 复位棋盘
  chessBoard.stop()
  chessBoard.start(saved.selfColor, false)
  // 逐步回放
  for (let i = 0; i < n; i++) {
    const step = hist.value[i]
    channel.emit('NET:CHESS:MOVE', { from: step.from, to: step.to })
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
  if (index.value <= 0) return
  index.value--
  applySteps(index.value)
}

function onNext() {
  const max = hist.value.length
  if (index.value >= max) return
  index.value++
  applySteps(index.value)
}
function handlePopState(_event: PopStateEvent) {
  window.history.pushState(null, '', window.location.href)
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
  window.history.pushState(null, '', window.location.href)
  // 监听 popstate 事件，防止后退操作
  window.addEventListener('popstate', (event) => {
    handlePopState(event)
  })
})

onUnmounted(() => {
  chessBoard?.stop()
  window.removeEventListener('popstate', handlePopState)
})
</script>

<template>
  <div class="h-full w-full flex flex-col items-center justify-start">
    <div class="relative mt-6">
      <canvas ref="background" />
      <canvas ref="chesses" class="absolute left-0 top-0" />
    </div>
    <div class="mt-4 flex gap-2">
      <button class="rounded bg-gray-600 px-4 py-2 text-white" @click="onExit">退出</button>
      <button class="rounded bg-blue-600 px-4 py-2 text-white" @click="onRestart">重新开始</button>
      <button :disabled="index <= 0" class="rounded bg-yellow-600 px-4 py-2 text-white disabled:opacity-40" @click="onPrev">上一步</button>
      <button :disabled="index >= hist.length" class="rounded bg-green-600 px-4 py-2 text-white disabled:opacity-40" @click="onNext">下一步</button>
    </div>
    <div class="mt-2 text-sm text-gray-700">第 {{ index }} 步 / 共 {{ hist.length }} 步</div>
  </div>
</template>
