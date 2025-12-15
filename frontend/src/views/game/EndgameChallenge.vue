<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, inject, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ChessBoard from '@/composables/ChessBoard'
import type { EndgameScenario } from '@/data/endgameData'
import { getScenarioById, validateScenario, quickFeasibilityCheck } from '@/data/endgameData'
import channel from '@/utils/channel'
import { showMsg } from '@/components/MessageBox'
import GameEndModal from '@/components/GameEndModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { useUserStore } from '@/store/useStore'

declare const window: any

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isPC = inject('isPC') as Ref<boolean>

const background = useTemplateRef('background')
const chesses = useTemplateRef('chesses')

let chessBoard: ChessBoard | null = null

const scenarioId = ref<string>((route.query.id as string) || '')
const status = ref<'idle' | 'playing' | 'win' | 'lose'>('idle')
const aiThinking = ref(false)
const redMovesUsed = ref(0)
const lastMove = ref('无')
const currentTurn = ref('—')
const resultModalVisible = ref(false)
const resultState = ref<'win' | 'lose' | 'draw' | null>(null)
const exitConfirmVisible = ref(false)
const endMessage = ref('')
const replayMode = ref(false)
const replayIndex = ref(0)

const moveStack = ref<Array<'red' | 'black'>>([])
const moveHistory = ref<Array<{ from: any; to: any; pieceName?: string; pieceColor?: string }>>([])
const progressSaved = ref(false)

const progressKey = 'endgame-progress'
interface ScenarioProgress {
  result?: 'win' | 'lose'
  attempts: number
  bestSteps?: number
}
const progress = ref<Record<string, ScenarioProgress>>({})

const activeScenario = computed<EndgameScenario | null>(() => getScenarioById(scenarioId.value))

const gridSize = computed(() => (isPC?.value ? 70 : 40))
const remainingMoves = computed(() => {
  const max = activeScenario.value?.maxMoves
  if (!max) return null
  return Math.max(0, max - redMovesUsed.value)
})
const aiLabel = computed(() => {
  const lvl = activeScenario.value?.aiLevel || 6
  if (lvl <= 2) return '简单'
  if (lvl <= 4) return '中等'
  return '困难'
})

function loadProgress() {
  try {
    const raw = localStorage.getItem(progressKey)
    if (!raw) return
    progress.value = JSON.parse(raw)
  } catch (e) {
    console.warn('读取残局进度失败', e)
  }
}

function saveProgress(id: string, result: 'win' | 'lose', steps?: number) {
  const existing = progress.value[id] || { attempts: 0 }
  const newProgress: ScenarioProgress = {
    result,
    attempts: existing.attempts + 1,
    bestSteps: existing.bestSteps,
  }
  if (result === 'win' && steps !== undefined) {
    if (!newProgress.bestSteps || steps < newProgress.bestSteps) {
      newProgress.bestSteps = steps
    }
  }
  progress.value[id] = newProgress
  try {
    localStorage.setItem(progressKey, JSON.stringify(progress.value))
  } catch (e) {
    console.warn('保存残局进度失败', e)
  }
}

function formatMoveLabel(from: any, to: any, pieceName?: string, pieceColor?: string) {
  const chineseNums = ['零', '一', '二', '三', '四', '五', '六', '七', '八', '九']
  const nameMap: Record<string, string> = {
    Horse: '马',
    馬: '马',
    马: '马',
    傌: '马',
    n: '马',
    Rook: '车',
    車: '车',
    车: '车',
    r: '车',
    俥: '车',
    Cannon: '炮',
    炮: '炮',
    砲: '炮',
    c: '炮',
    Bishop: pieceColor === 'red' ? '相' : '象',
    相: '相',
    象: '象',
    b: '相',
    Advisor: pieceColor === 'red' ? '仕' : '士',
    仕: '仕',
    士: '士',
    a: '仕',
    King: pieceColor === 'red' ? '帅' : '将',
    帥: '帅',
    将: '将',
    將: '将',
    k: '将',
    Pawn: pieceColor === 'red' ? '兵' : '卒',
    兵: '兵',
    卒: '卒',
    s: '兵',
  }
  const pieceChar = pieceName ? nameMap[pieceName] || '棋' : '棋'
  const fromFile = pieceColor === 'red' ? 9 - from.x : from.x + 1
  const toFile = pieceColor === 'red' ? 9 - to.x : to.x + 1
  const forward = pieceColor === 'red' ? to.y < from.y : to.y > from.y
  const action = from.x === to.x ? '平' : forward ? '进' : '退'
  const fromLabel = chineseNums[fromFile] || String(fromFile)
  const toLabel = chineseNums[toFile] || String(toFile)
  return `${pieceChar}${fromLabel}${action}${toLabel}`
}

function initBoard() {
  const canvasBackground = background.value as HTMLCanvasElement
  const canvasChesses = chesses.value as HTMLCanvasElement
  const ctxBackground = canvasBackground?.getContext('2d')
  const ctxChesses = canvasChesses?.getContext('2d')
  if (!ctxBackground || !ctxChesses) {
    throw new Error('画布初始化失败')
  }
  chessBoard = new ChessBoard(canvasBackground, canvasChesses, gridSize.value)
}

function startScenario(options?: { keepHistory?: boolean }) {
  const keepHistory = options?.keepHistory ?? false
  const scenario = activeScenario.value
  if (!scenario || !chessBoard) {
    showMsg('请先选择一个残局关卡')
    router.push('/endgame')
    return
  }

  const valid = validateScenario(scenario)
  if (!valid.valid) {
    showMsg(valid.errors.join('；'))
    return
  }
  const feasibility = quickFeasibilityCheck(scenario)
  if (!feasibility.playable) {
    showMsg(`布局存在问题：${feasibility.issues.join('；')}`)
    return
  }

  const res = chessBoard.startEndgame({
    pieces: scenario.initialPieces,
    currentTurn: scenario.initialTurn,
    selfColor: 'red',
    aiMode: true,
  })
  if (!res?.success) return

  status.value = 'playing'
  redMovesUsed.value = 0
  moveStack.value = []
  if (!keepHistory) moveHistory.value = []
  lastMove.value = '无'
  resultState.value = null
  resultModalVisible.value = false
  replayMode.value = false
  replayIndex.value = 0
  endMessage.value = ''
  progressSaved.value = false
  currentTurn.value = chessBoard.currentRole === 'self' ? '你的回合' : '对手回合'
  // 残局挑战固定红方先手，无需AI先走
}

function handleWin() {
  status.value = 'win'
  resultState.value = 'win'
  resultModalVisible.value = true
  const totalSteps = redMovesUsed.value
  if (!progressSaved.value) {
    saveProgress(activeScenario.value?.id || '', 'win', totalSteps)
    progressSaved.value = true
  }
  chessBoard?.disableInteraction()
  endMessage.value = '挑战成功！'
  showMsg('挑战成功！')
}

function startReplay() {
  if (!chessBoard) return
  if (!moveHistory.value.length) {
    showMsg('暂无可复盘的步数')
    return
  }
  replayMode.value = true
  replayIndex.value = 0
  resultModalVisible.value = false
  applyReplaySteps(0)
}

function applyReplaySteps(n: number) {
  if (!chessBoard) return
  const scenario = activeScenario.value
  if (!scenario) return
  
  chessBoard.startEndgame({
    pieces: scenario.initialPieces,
    currentTurn: scenario.initialTurn,
    selfColor: 'red',
    aiMode: false,
  })
  
  for (let i = 0; i < n && i < moveHistory.value.length; i++) {
    const step = moveHistory.value[i]
    chessBoard.replayMove(step.from, step.to)
  }
  chessBoard.disableInteraction()
  status.value = 'idle'
  endMessage.value = ''
}

function exitReplay() {
  replayMode.value = false
  replayIndex.value = 0
  resultModalVisible.value = true
}

function replayRestart() {
  replayIndex.value = 0
  applyReplaySteps(0)
}

function replayPrev() {
  if (replayIndex.value <= 0) return
  replayIndex.value--
  applyReplaySteps(replayIndex.value)
}

function replayNext() {
  if (replayIndex.value >= moveHistory.value.length) return
  replayIndex.value++
  applyReplaySteps(replayIndex.value)
}

function handleLose(reason?: string) {
  status.value = 'lose'
  resultState.value = 'lose'
  resultModalVisible.value = true
  chessBoard?.disableInteraction()
  endMessage.value = reason || '对局结束'
  if (reason) showMsg(reason)
  if (!progressSaved.value) {
    saveProgress(activeScenario.value?.id || '', 'lose')
    progressSaved.value = true
  }
}

function requestAIMove() {
  const scenario = activeScenario.value
  if (!scenario || !chessBoard) return
  if (!window.logic || !window.logic.getAIMoveAdaptive) {
    showMsg('AI 引擎未加载')
    return
  }
  aiThinking.value = true
  try {
    const board = chessBoard.getCurrentBoard()
    const aiColor = 'black'
    const moveData = window.logic.getAIMoveAdaptive(board, aiColor, {
      playerSkill: scenario.aiLevel || 6,
      useOpeningBook: false,
      moveNumber: chessBoard.stepsNum / 2,
    })
    if (!moveData) {
      aiThinking.value = false
      if (window.logic.isCheckmate(board, aiColor)) {
        channel.emit('LOCAL:GAME:END', { winner: 'red' })
      }
      return
    }
    let fromX = moveData.from[1]
    let fromY = moveData.from[0]
    let toX = moveData.to[1]
    let toY = moveData.to[0]
    const fromPos = { x: fromX, y: fromY }
    const toPos = { x: toX, y: toY }
    let applied = chessBoard.applyAIMove(fromPos, toPos)
    if (!applied) {
      try {
        const newBoard = window.logic.applyMove(board, moveData.from, moveData.to)
        chessBoard.setCurrentBoard(newBoard)
        chessBoard.render()
        applied = true
      } catch (e) {
        console.error('AI fallback move failed', e)
      }
    }
    if (!applied) {
      aiThinking.value = false
      return
    }
    chessBoard.render()
    const newBoard = chessBoard.getCurrentBoard()
    const playerCheckmated = window.logic.isCheckmate(newBoard, 'red')
    if (playerCheckmated) {
      channel.emit('LOCAL:GAME:END', { winner: 'black' })
      aiThinking.value = false
      return
    }
    if (window.logic.isInCheck(newBoard, 'red')) {
      showMsg('你被将军了！')
    }
    chessBoard.setCurrentRole('self')
  } finally {
    aiThinking.value = false
  }
}

function regret() {
  if (!chessBoard) return
  if (status.value !== 'playing') {
    showMsg('当前不允许悔棋')
    return
  }
  const undone = chessBoard.regretLastTurn()
  if (!undone) {
    showMsg('没有可悔的步数')
    return
  }
  for (let i = 0; i < undone; i++) {
    const popped = moveStack.value.pop()
    if (popped === 'red') {
      redMovesUsed.value = Math.max(0, redMovesUsed.value - 1)
    }
  }
  lastMove.value = '无'
}

function handleBoardMove(payload: any) {
  const { from, to, pieceName, pieceColor } = payload || {}
  if (!pieceColor) return
  moveStack.value.push(pieceColor)
  moveHistory.value.push({ from, to, pieceName, pieceColor })
  if (pieceColor === 'red') {
    redMovesUsed.value += 1
    const max = activeScenario.value?.maxMoves
    if (max && redMovesUsed.value > max && status.value === 'playing') {
      handleLose('步数用尽')
      return
    }
  }
  lastMove.value = formatMoveLabel(from, to, pieceName, pieceColor)
}

function bindEvents() {
  channel.on('LOCAL:GAME:END', (payload: any) => {
    const winner = payload?.winner
    if (winner === 'red') handleWin()
    else handleLose()
  })
  channel.on('BOARD:ROLE:CHANGE', ({ currentRole }: any) => {
    currentTurn.value = currentRole === 'self' ? '你的回合' : '对手回合'
  })
  ;(channel as any).on('BOARD:MOVE:MADE', handleBoardMove)
  channel.on('GAME:MOVE', () => {
    if (status.value !== 'playing') return
    chessBoard?.setCurrentRole('enemy')
    setTimeout(() => requestAIMove(), 400)
  })
}

function unbindEvents() {
  channel.off('LOCAL:GAME:END')
  channel.off('BOARD:ROLE:CHANGE')
  ;(channel as any).off('BOARD:MOVE:MADE', handleBoardMove as any)
  channel.off('GAME:MOVE')
}

function exitChallenge() {
  if (status.value === 'playing') {
    exitConfirmVisible.value = true
    return
  }
  router.push('/endgame')
}

function confirmExit() {
  exitConfirmVisible.value = false
  router.push('/endgame')
}

function cancelExit() {
  exitConfirmVisible.value = false
}

onMounted(() => {
  loadProgress()
  if (!scenarioId.value) {
    router.push('/endgame')
    return
  }
  initBoard()
  bindEvents()
  startScenario()
})

watch(
  () => route.query.id,
  (id) => {
    if (typeof id === 'string' && id) {
      scenarioId.value = id
      startScenario()
    }
  },
)

onUnmounted(() => {
  unbindEvents()
  chessBoard?.stop()
})
</script>

<template>
  <div class="h-full w-full flex flex-col sm:flex-row">
    <div class="block h-1/5 sm:h-full flex-1" />
    <div class="relative h-3/5 w-full sm:h-full sm:w-5/12 flex flex-col">
      <div class="relative flex-1 w-full">
        <!-- AI思考提示 -->
        <div v-if="aiThinking" class="absolute top-4 left-1/2 -translate-x-1/2 z-10 bg-yellow-100 text-yellow-800 px-4 py-2 rounded-lg font-semibold">
          AI思考中...
        </div>
        <canvas
          ref="background"
          class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2"
        />
        <canvas ref="chesses" class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />
      </div>
      <div class="flex justify-center space-x-4 mb-20">
        <button
          class="border-0 rounded-2xl bg-yellow-500 text-white p-4 transition-all duration-200"
          text="xl"
          hover="bg-yellow-600"
          @click="regret"
        >
          悔棋
        </button>
        <button
          class="border-0 rounded-2xl bg-gray-500 text-white p-4 transition-all duration-200"
          text="xl"
          hover="bg-gray-600"
          @click="exitChallenge"
        >
          退出
        </button>
      </div>
    </div>
    <div class="sm:h-full flex-1 flex flex-col pt-12 pb-20 pr-48">
      <!-- 残局对战信息面板 -->
      <div class="bg-white/80 backdrop-blur rounded-xl shadow-sm p-4 mb-4 flex flex-col border border-gray-200">
        <div class="flex items-center justify-between w-full mb-4">
          <div class="flex flex-col items-center w-1/3">
            <img :src="userStore.userInfo?.avatar || '/images/default_avatar.png'" alt="玩家头像" class="w-12 h-12 rounded-full mb-1 object-cover border-2 border-red-500" />
            <span class="text-xs truncate w-full text-center font-medium">{{ userStore.userInfo?.name }}</span>
            <span class="text-xs font-bold mt-1 text-red-600">红方</span>
          </div>
          <div class="text-2xl font-black text-gray-400 italic mx-2">VS</div>
          <div class="flex flex-col items-center w-1/3">
            <div class="w-12 h-12 rounded-full mb-1 bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center text-white font-bold text-lg">
              🤖
            </div>
            <span class="text-xs truncate w-full text-center font-medium">电脑</span>
            <span class="text-xs font-bold mt-1 text-black">黑方</span>
          </div>
        </div>
        <div class="flex justify-between w-full space-x-2">
          <div class="w-1/2 text-xs space-y-1 bg-gray-50 p-2 rounded-lg">
            <div class="flex justify-between items-center">
              <span class="text-gray-500">昵称</span>
              <span class="font-bold text-gray-700">{{ userStore.userInfo?.name || '玩家' }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">经验</span>
              <span class="font-bold text-gray-700">{{ userStore.userInfo?.exp || 0 }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">胜率</span>
              <span class="font-bold text-gray-700">{{ (userStore.userInfo?.winRate || 0).toFixed(2) }}%</span>
            </div>
          </div>
          <div class="w-1/2 text-xs space-y-1 bg-gray-50 p-2 rounded-lg">
            <div class="flex justify-between items-center">
              <span class="text-gray-500">AI难度</span>
              <span class="font-bold text-gray-700">{{ aiLabel }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">残局</span>
              <span class="font-bold text-gray-700">{{ activeScenario?.name || '未知' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 剩余步数提醒 -->
      <div v-if="status === 'playing' && activeScenario?.maxMoves && !replayMode" class="bg-orange-50 border border-orange-300 rounded-lg p-3 mb-4">
        <div class="text-center">
          <span class="text-sm text-orange-700">剩余步数：</span>
          <span class="text-2xl font-bold" :class="remainingMoves && remainingMoves <= 3 ? 'text-red-600' : 'text-orange-600'">{{ remainingMoves }}</span>
        </div>
      </div>

      <!-- 复盘控制面板 -->
      <div v-if="replayMode" class="bg-purple-50 border border-purple-300 rounded-lg p-3 mb-4">
        <div class="text-sm font-semibold text-purple-900 mb-2 text-center">复盘模式</div>
        <div class="flex gap-2 justify-center mb-2">
          <button class="rounded bg-gray-600 px-3 py-1 text-white text-sm" @click="exitReplay">退出复盘</button>
          <button class="rounded bg-blue-600 px-3 py-1 text-white text-sm" @click="replayRestart">重新开始</button>
        </div>
        <div class="flex gap-2 justify-center mb-2">
          <button :disabled="replayIndex <= 0" class="rounded bg-yellow-600 px-3 py-1 text-white text-sm disabled:opacity-40" @click="replayPrev">上一步</button>
          <button :disabled="replayIndex >= moveHistory.length" class="rounded bg-green-600 px-3 py-1 text-white text-sm disabled:opacity-40" @click="replayNext">下一步</button>
        </div>
        <div class="text-xs text-center text-gray-700">第 {{ replayIndex }} 步 / 共 {{ moveHistory.length }} 步</div>
      </div>

      <!-- 游戏统计 -->
      <div v-if="!replayMode" class="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
        <div class="text-sm font-semibold text-blue-900 mb-2">游戏信息</div>
        <div class="grid grid-cols-2 gap-2 text-xs">
          <div class="flex justify-between">
            <span class="text-gray-600">已走步数:</span>
            <span class="font-bold">{{ chessBoard?.stepsNum || 0 }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">游戏模式:</span>
            <span class="font-bold">残局挑战</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">当前回合:</span>
            <span class="font-bold">{{ currentTurn }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">最近落子:</span>
            <span class="font-bold">{{ lastMove }}</span>
          </div>
        </div>
      </div>

    </div>
  </div>
  <GameEndModal
    :visible="resultModalVisible"
    :result="resultState"
    :on-review="startReplay"
    :on-quit="exitChallenge"
    :message-override="endMessage"
    @close="resultModalVisible = false"
  />
  <ConfirmModal
    :visible="exitConfirmVisible"
    title="确定要退出当前残局？"
    message="退出后将返回关卡列表，当前进度不会保存。"
    confirmText="退出"
    cancelText="取消"
    :on-confirm="confirmExit"
    :on-cancel="cancelExit"
  />
</template>
