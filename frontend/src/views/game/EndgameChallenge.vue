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
import { getGameState, clearGameState } from '@/store/gameStore'
import { getEndgameProgress, saveEndgameProgress } from '@/api/endgame/progress'
import { reportEndgameComplete } from '@/api/endgame/complete'
import { getProfile } from '@/api/user/getProfile'

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

const getProgressKey = () => {
  const userId = userStore.userInfo?.id || 'guest'
  return `endgame-progress-${userId}`
}
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
  // 优先从后端获取进度，保证多设备同步；失败时回退到本地 localStorage
  const userId = userStore.userInfo?.id
  if (!userId) {
    // 未登录用户直接走本地缓存
    try {
      const raw = localStorage.getItem(getProgressKey())
      if (!raw) return
      progress.value = JSON.parse(raw)
    } catch (e) {
      console.warn('读取残局进度失败', e)
    }
    return
  }

  getEndgameProgress()
    .then((resp: any) => {
      const data = resp && typeof resp === 'object' && 'data' in resp ? resp.data : resp
      if (!data || !Array.isArray(data.progress)) return
      const map: Record<string, ScenarioProgress> = {}
      data.progress.forEach((item: any) => {
        const id = String(item.scenario_id)
        const attempts = Number(item.attempts) || 0
        const bestSteps = typeof item.best_steps === 'number' ? item.best_steps : undefined
        const lastResult = item.last_result === 'win' || item.last_result === 'lose' ? item.last_result : undefined
        map[id] = {
          attempts,
          bestSteps,
          result: lastResult,
        }
      })
      progress.value = map
      // 同步一份到本地缓存，做离线降级
      try {
        localStorage.setItem(getProgressKey(), JSON.stringify(progress.value))
      } catch {}
    })
    .catch((e: any) => {
      console.warn('从后端获取残局进度失败，回退到本地缓存', e)
      try {
        const raw = localStorage.getItem(getProgressKey())
        if (!raw) return
        progress.value = JSON.parse(raw)
      } catch (err) {
        console.warn('读取本地残局进度失败', err)
      }
    })
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
  // 先更新本地缓存，保证界面即时一致
  try {
    localStorage.setItem(getProgressKey(), JSON.stringify(progress.value))
  } catch (e) {
    console.warn('保存本地残局进度失败', e)
  }

  // 若已登录，则同步到后端，保证不同浏览器/设备一致
  const userId = userStore.userInfo?.id
  if (userId) {
    saveEndgameProgress({ scenarioId: id, result, steps })
      .catch((e: any) => {
        console.warn('同步残局进度到后端失败，将使用本地缓存为准', e)
      })
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
  const id = activeScenario.value?.id || ''
  const diff = activeScenario.value?.difficulty || ''
  const wasCompleted = id ? (progress.value?.[id]?.result === 'win') : false
  // 首次通关经验奖励：高级+100 / 中级+50 / 其它+0（只发放一次）
  try {
    if (!wasCompleted && id) {
      reportEndgameComplete({ scenarioId: id, difficulty: diff as any, success: true })
        .then(async (resp: any) => {
          // 刷新资料以更新经验
          try {
            const prof = await getProfile()
            const d = prof && typeof prof === 'object' && 'data' in prof ? (prof as any).data : prof
            if (d) userStore.setUser(d)
          } catch {}
        })
        .catch(() => {})
    }
  } catch {}
  if (!progressSaved.value) {
    saveProgress(activeScenario.value?.id || '', 'win', totalSteps)
    progressSaved.value = true
  }
  // 对局已结束，清除残局对局的临时局面，防止刷新后继续在结束局面上走棋
  clearGameState()
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
  replayIndex.value = moveHistory.value.length
  resultModalVisible.value = false
  applyReplaySteps(moveHistory.value.length)
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
  // 失败同样清除局面缓存，刷新后重新开局
  clearGameState()
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
  // 主动退出残局时清除局面缓存
  clearGameState()
  router.push('/endgame')
}

function confirmExit() {
  exitConfirmVisible.value = false
  // 中途主动确认退出当前残局，对局局面不再保留，刷新或重进时从初始布局开始
  clearGameState()
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
  // 尝试从会话中恢复残局局面
  const saved = getGameState()
  if (saved && saved.mode === 'endgame' && saved.moveHistory && saved.moveHistory.length > 0) {
    try {
      chessBoard?.restoreState(saved as any)
      status.value = 'playing'
      // 基于保存的历史恢复本页统计信息
      moveHistory.value = saved.moveHistory.map((step: any) => ({
        from: step.from,
        to: step.to,
        pieceName: step.pieceName,
        pieceColor: step.pieceColor,
      }))
      moveStack.value = []
      redMovesUsed.value = 0
      saved.moveHistory.forEach((step: any) => {
        if (step.pieceColor === 'red' || step.pieceColor === 'black') {
          moveStack.value.push(step.pieceColor)
        }
        if (step.pieceColor === 'red') {
          redMovesUsed.value += 1
        }
      })
      const last = moveHistory.value[moveHistory.value.length - 1]
      lastMove.value = last ? formatMoveLabel(last.from, last.to, last.pieceName, last.pieceColor) : '无'
      currentTurn.value = chessBoard && chessBoard.currentRole === 'self' ? '你的回合' : '对手回合'
    } catch (e) {
      console.warn('恢复残局局面失败，重新开始关卡', e)
      startScenario()
    }
  } else {
    startScenario()
  }
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
  // 若当前不在进行对局（已结束或尚未开始），卸载时清理残局局面缓存，
  // 防止之后进入其它模式（如本地对战）时误用残局存档
  if (status.value !== 'playing') {
    clearGameState()
  }
})
</script>

<template>
  <div class="h-full w-full bg-[#fdf6e3] flex flex-col sm:flex-row relative overflow-hidden">
    <!-- 背景装饰 -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute -top-[20%] -left-[10%] w-[70%] h-[70%] rounded-full bg-amber-200/20 blur-3xl"></div>
      <div class="absolute top-[40%] -right-[10%] w-[60%] h-[60%] rounded-full bg-orange-200/20 blur-3xl"></div>
    </div>

    <!-- 主布局容器 -->
    <div class="relative z-10 flex-1 flex flex-col sm:flex-row h-full max-w-[1200px] mx-auto w-full p-2 sm:p-4 gap-4 justify-center items-center">

      <!-- 左侧/中间：棋盘区域 -->
      <div class="flex-none flex flex-col items-center justify-center">
        <!-- 棋盘容器 -->
        <div class="relative w-[90vw] sm:w-[650px] aspect-[9/10] flex-none flex items-center justify-center">
          <!-- 棋盘背景装饰 -->
          <div class="absolute inset-4 bg-[#eecfa1] rounded shadow-2xl transform rotate-0 opacity-50 blur-sm"></div>

          <!-- AI思考提示 -->
          <div v-if="aiThinking" class="absolute top-4 left-1/2 -translate-x-1/2 z-20 bg-amber-100/90 backdrop-blur text-amber-800 px-6 py-2 rounded-full font-bold shadow-lg border border-amber-200 flex items-center gap-2 animate-pulse">
            <svg class="animate-spin h-4 w-4 text-amber-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            AI思考中...
          </div>

          <canvas
            ref="background"
            class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 shadow-2xl rounded-lg"
          />
          <canvas ref="chesses" class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 pointer-events-auto" />
        </div>

        <!-- 底部按钮栏 -->
        <div class="mt-6 flex flex-wrap justify-center gap-3 sm:gap-6 w-full max-w-[600px]">
          <button
            class="group relative px-6 py-2.5 bg-amber-100 text-amber-900 rounded-xl font-bold shadow-sm hover:bg-amber-200 hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2 border border-amber-200"
            @click="regret"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
            </svg>
            悔棋
          </button>
          <button
            class="group relative px-6 py-2.5 bg-gray-100 text-gray-700 rounded-xl font-bold shadow-sm hover:bg-gray-200 hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2 border border-gray-200"
            @click="exitChallenge"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            退出
          </button>
        </div>
      </div>

      <!-- 右侧：信息面板 -->
      <div class="w-full sm:w-72 lg:w-80 flex-none flex flex-col gap-4 h-auto sm:h-full overflow-y-auto">
        <!-- 残局对战信息面板 -->
        <div class="bg-white/60 backdrop-blur-md rounded-2xl shadow-sm p-4 border border-white/50 flex flex-col">
          <div class="flex items-center justify-between w-full mb-4">
            <!-- 玩家 -->
            <div class="flex flex-col items-center w-1/3 group">
              <div class="relative">
                <img :src="userStore.userInfo?.avatar || '/images/default_avatar.png'" alt="玩家头像" class="w-14 h-14 rounded-full mb-2 object-cover border-4 border-red-500 shadow-md transition-transform group-hover:scale-105" />
                <div class="absolute -bottom-1 -right-1 w-6 h-6 rounded-full bg-red-500 flex items-center justify-center text-xs font-bold text-white shadow-sm">
                  红
                </div>
              </div>
              <span class="text-sm truncate w-full text-center font-bold text-amber-900">{{ userStore.userInfo?.name }}</span>
            </div>

            <!-- VS -->
            <div class="flex flex-col items-center justify-center">
              <span class="text-3xl font-black text-amber-200/80 italic">VS</span>
            </div>

            <!-- 电脑 -->
            <div class="flex flex-col items-center w-1/3 group">
              <div class="relative">
                <div class="w-14 h-14 rounded-full mb-2 bg-gradient-to-br from-indigo-500 to-blue-600 flex items-center justify-center text-white font-bold text-2xl shadow-md border-4 border-white transition-transform group-hover:scale-105">
                  🤖
                </div>
                <div class="absolute -bottom-1 -right-1 w-6 h-6 rounded-full bg-gray-800 flex items-center justify-center text-xs font-bold text-white shadow-sm">
                  黑
                </div>
              </div>
              <span class="text-sm truncate w-full text-center font-bold text-amber-900">电脑</span>
            </div>
          </div>

          <!-- 数据展示 -->
          <div class="flex justify-between w-full gap-2">
            <div class="flex-1 bg-amber-50/50 p-2 rounded-lg border border-amber-100">
              <div class="flex justify-between items-center text-xs mb-1">
                <span class="text-amber-800/60">经验</span>
                <span class="font-bold text-amber-900">{{ userStore.userInfo?.exp || 0 }}</span>
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="text-amber-800/60">胜率</span>
                <span class="font-bold text-amber-900">{{ (userStore.userInfo?.winRate || 0).toFixed(0) }}%</span>
              </div>
            </div>
            <div class="flex-1 bg-indigo-50/50 p-2 rounded-lg border border-indigo-100">
              <div class="flex justify-between items-center text-xs mb-1">
                <span class="text-indigo-800/60">难度</span>
                <span class="font-bold text-indigo-900">{{ aiLabel }}</span>
              </div>
              <div class="flex justify-between items-center text-xs">
                <span class="text-indigo-800/60">残局</span>
                <span class="font-bold text-indigo-900 truncate max-w-[4em]">{{ activeScenario?.name || '未知' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 剩余步数提醒 -->
        <div v-if="status === 'playing' && activeScenario?.maxMoves && !replayMode" class="bg-orange-50/80 backdrop-blur border border-orange-200 rounded-2xl p-4 shadow-sm flex items-center justify-between">
          <span class="text-sm font-bold text-orange-800">剩余步数</span>
          <span class="text-3xl font-black" :class="remainingMoves && remainingMoves <= 3 ? 'text-red-600 animate-pulse' : 'text-orange-600'">{{ remainingMoves }}</span>
        </div>

        <!-- 复盘控制面板 -->
        <div v-if="replayMode" class="bg-purple-50/80 backdrop-blur border border-purple-200 rounded-2xl p-4 shadow-sm">
          <div class="text-sm font-bold text-purple-900 mb-3 text-center flex items-center justify-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            复盘模式
          </div>
          <div class="flex gap-2 justify-center mb-3">
            <button class="rounded-lg bg-blue-600 hover:bg-blue-700 text-white px-4 py-1.5 text-sm font-bold shadow-sm transition-colors" @click="replayRestart">重新开始</button>
          </div>
          <div class="flex gap-2 justify-center mb-3">
            <button :disabled="replayIndex <= 0" class="rounded-lg bg-amber-500 hover:bg-amber-600 text-white px-4 py-1.5 text-sm font-bold shadow-sm disabled:opacity-40 disabled:cursor-not-allowed transition-colors" @click="replayPrev">上一步</button>
            <button :disabled="replayIndex >= moveHistory.length" class="rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-1.5 text-sm font-bold shadow-sm disabled:opacity-40 disabled:cursor-not-allowed transition-colors" @click="replayNext">下一步</button>
          </div>
          <div class="text-xs text-center font-medium text-purple-800/70">第 {{ replayIndex }} 步 / 共 {{ moveHistory.length }} 步</div>
        </div>

        <!-- 游戏统计 -->
        <div v-if="!replayMode" class="bg-white/60 backdrop-blur-md border border-white/50 rounded-2xl p-4 shadow-sm">
          <div class="flex items-center gap-2 mb-3">
            <div class="w-1 h-4 bg-amber-500 rounded-full"></div>
            <span class="text-sm font-bold text-amber-900">对局信息</span>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="bg-amber-50 rounded-xl p-3 border border-amber-100 flex flex-col items-center justify-center">
              <span class="text-xs text-amber-800/60 mb-1">已走步数</span>
              <span class="font-mono font-bold text-amber-900 text-lg">{{ chessBoard?.stepsNum || 0 }}</span>
            </div>
            <div class="bg-amber-50 rounded-xl p-3 border border-amber-100 flex flex-col items-center justify-center">
              <span class="text-xs text-amber-800/60 mb-1">游戏模式</span>
              <span class="font-bold text-amber-900 text-lg">残局挑战</span>
            </div>
            <div class="bg-amber-50 rounded-xl p-3 border border-amber-100 flex flex-col items-center justify-center">
              <span class="text-xs text-amber-800/60 mb-1">当前回合</span>
              <span class="font-bold text-amber-900 text-lg">{{ currentTurn }}</span>
            </div>
            <div class="bg-amber-50 rounded-xl p-3 border border-amber-100 flex flex-col items-center justify-center">
              <span class="text-xs text-amber-800/60 mb-1">最近落子</span>
              <span class="font-mono font-bold text-amber-900 text-lg">{{ lastMove }}</span>
            </div>
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
