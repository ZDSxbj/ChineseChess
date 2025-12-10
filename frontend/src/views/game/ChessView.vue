<script lang="ts" setup>
import type { Ref } from 'vue'
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { useRouter } from 'vue-router'
import ChatPanel from '@/components/ChatPanel.vue'
import GameEndModal from '@/components/GameEndModal.vue'
import { showMsg } from '@/components/MessageBox'
import RegretModal from '@/components/RegretModal.vue'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState, getGameState, saveGameState, saveModalState, getModalState, clearModalState } from '@/store/gameStore'
import { useUserStore } from '@/store/useStore'
import channel from '@/utils/channel'

const router = useRouter()
const userStore = useUserStore()
const chatPanelRef = ref()
const networkPlay = ref(false)
const opponentInfo = ref<any>(null)
const selfColor = ref<'red' | 'black'>('red')

const background = useTemplateRef('background')
const chesses = useTemplateRef('chesses')

const isPC = inject('isPC') as Ref<boolean>

let chessBoard: ChessBoard
const ws = inject('ws') as WebSocketService

const regretModalVisible = ref(false)
const regretModalType = ref<'requesting' | 'responding'>('requesting')
const gameOver = ref(false)
const drawModalVisible = ref(false)
const drawModalType = ref<'requesting' | 'responding'>('requesting')
const endModalVisible = ref(false)
const endResult = ref<'win' | 'lose' | 'draw' | null>(null)
const resignConfirmVisible = ref(false)
const quitConfirmVisible = ref(false)
const pendingQuitAfterResign = ref(false)

function handleRegretAccept() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendRegretResponse(true)
  const steps = chessBoard.isMyTurn() ? 1 : 2
  for (let i = 0; i < steps; i++) {
    chessBoard?.regretMove()
  }
  showMsg(`悔了${steps}步棋`)
  chessBoard?.setCurrentRole('enemy')
  regretModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
}

function handleRegretReject() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendRegretResponse(false)
  regretModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
}

function handleDrawAccept() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawResponse(true)
  // 接受和棋，后端会广播 GAME:END
  drawModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
}

function handleDrawReject() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawResponse(false)
  drawModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
  showMsg('已拒绝对方的和棋请求')
}

function review() {
  // 确保保存最新的局面到 sessionStorage
  saveGameState({
    isNetPlay: chessBoard.isNetworkPlay(),
    selfColor: chessBoard.SelfColor,
    moveHistory: chessBoard.moveHistoryList,
    currentRole: chessBoard.currentRole,
    opponentInfo: opponentInfo.value,
  })
  clearModalState() // 退出时清除模态状态，防止下次进入时残留
  router.push('/game/replay')
}

function decideSize(isPCBool: boolean) {
  return isPCBool ? 70 : 40
}

watch(isPC, (newIsPC) => {
  const gridSize = decideSize(newIsPC)
  chessBoard?.redraw(gridSize)
})

function confirmGiveUp() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  if (networkPlay.value) {
    resignConfirmVisible.value = true
  }
}

function handleGiveUpConfirm() {
  resignConfirmVisible.value = false
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  if (networkPlay.value) {
    ws?.giveUp()
  }
  else {
    const opponentColor = chessBoard.SelfColor === 'red' ? 'black' : 'red'
    channel.emit('GAME:END', { winner: opponentColor, online: false })
  }
}

function handleGiveUpCancel() {
  resignConfirmVisible.value = false
}

function offerDraw() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawRequest()
  drawModalType.value = 'requesting'
  drawModalVisible.value = true
  saveModalState({
    regretModalVisible: regretModalVisible.value,
    regretModalType: regretModalType.value,
    drawModalVisible: drawModalVisible.value,
    drawModalType: drawModalType.value,
  })
  showMsg('已发送和棋请求')
}

function quit() {
  if (!gameOver.value) {
    quitConfirmVisible.value = true
    return
  }
  clearGameState()
  clearModalState()
  router.push('/')
}

function handleQuitConfirm() {
  quitConfirmVisible.value = false
  if (networkPlay.value) {
    ws?.giveUp()
  }
  clearGameState()
  clearModalState()
  router.push('/')
}

function handleQuitCancel() {
  quitConfirmVisible.value = false
}
// 新增悔棋函数
function regret() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  if (chessBoard) {
    if (!chessBoard.stepsNum) {
      showMsg('没有可悔棋的步数')
    }
    else if (chessBoard.isNetworkPlay()) {
      if (chessBoard.Color === 'black' && chessBoard.stepsNum === 1) {
        showMsg('没有可悔棋的步数')
        return
      }
      ws.sendRegretRequest()
      regretModalType.value = 'requesting'
      regretModalVisible.value = true
      saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
    }
    else {
      chessBoard.regretMove()
    }
  }
}

// 当前回合与最近一步显示（响应式）
const currentTurn = ref<string>('—')
const lastMove = ref<string>('无')

function formatMoveLabel(from: any, to: any, pieceName?: string, pieceColor?: string) {
  const chineseNums = ['零','一','二','三','四','五','六','七','八','九']
  const nameMap: Record<string,string> = {
    'Horse': '马', '馬': '马', '马': '马', '傌': '马', 'n': '马',
    'Rook': '车', '車': '车', '车': '车', 'r': '车', '俥': '车',
    'Cannon': '炮', '炮': '炮', 'c': '炮', '砲': '炮',
    // 象/相 根据棋子颜色显示不同字形
    'Bishop': pieceColor === 'red' ? '相' : '象', '相': '相', '象': '象', 'b': '相',
    // 士/仕 由颜色决定，黑方应显示“士”
    'Advisor': pieceColor === 'red' ? '仕' : '士', '仕': '仕', '士': '士', 'a': '仕',
    // 王/帅/将 兼容繁简及 engine 产生的字符
    'King': pieceColor === 'red' ? '帅' : '将', '帥': '帅', '帅': '帅', '將': '将', '将': '将', 'k': '将',
    // 兵/卒
    'Pawn': pieceColor === 'red' ? '兵' : '卒', '兵': '兵', '卒': '卒', 's': '兵',
  }
  const pieceChar = pieceName ? (nameMap[pieceName] || nameMap[pieceName as any] || '棋') : '棋'
  const fromFile = pieceColor === 'red' ? 9 - from.x : from.x + 1
  const toFile = pieceColor === 'red' ? 9 - to.x : to.x + 1
  let action = ''
  if (from.x === to.x) {
    action = '平'
  } else {
    const forward = pieceColor === 'red' ? to.y < from.y : to.y > from.y
    action = forward ? '进' : '退'
  }
  const fromLabel = chineseNums[fromFile] || String(fromFile)
  const toLabel = chineseNums[toFile] || String(toFile)
  return `${pieceChar}${fromLabel}${action}${toLabel}`
}

function handlePopState(_event: PopStateEvent) {
  window.history.pushState(null, '', window.location.href)
  showMsg('请通过应用内的导航按钮进行操作')
}

onMounted(() => {
  const gridSize = decideSize(isPC.value)
  const canvasBackground = background.value as HTMLCanvasElement
  const canvasChesses = chesses.value as HTMLCanvasElement
  const ctxBackground = canvasBackground.getContext('2d')
  const ctxChesses = canvasChesses.getContext('2d')

  if (!ctxBackground || !ctxChesses) {
    throw new Error('Failed to get canvas context')
  }

  chessBoard = new ChessBoard(canvasBackground, canvasChesses, gridSize)
  chessBoard.start('red', false)
  networkPlay.value = false
  console.log('ChessBoard started')
  // 尝试加载保存的游戏状态
  const savedState = getGameState()
  if (savedState) {
    chessBoard.restoreState(savedState)
    if (savedState.opponentInfo) {
      opponentInfo.value = savedState.opponentInfo
    }
    selfColor.value = chessBoard.SelfColor
    console.log('Game state restored from localStorage')
  }

  // 恢复弹窗状态（若有）
  const modalState = getModalState()
  if (modalState) {
    regretModalVisible.value = !!modalState.regretModalVisible
    regretModalType.value = modalState.regretModalType || 'requesting'
    drawModalVisible.value = !!modalState.drawModalVisible
    drawModalType.value = modalState.drawModalType || 'requesting'
    // 恢复结束模态（若之前为显示状态）
    endModalVisible.value = !!modalState.endModalVisible
    endResult.value = (modalState.endResult as any) || null
  }

  // 新增：保存初始游戏状态
  saveGameState({
    isNetPlay: chessBoard.isNetworkPlay(),
    selfColor: chessBoard.SelfColor,
    moveHistory: chessBoard.moveHistoryList, // 初始为空
    currentRole: chessBoard.currentRole,
    opponentInfo: opponentInfo.value,
  })
  // 清空消息队列
  channel.clearQueue('NET:GAME:END')
  // 初始化当前回合与最近落子显示
  try {
    currentTurn.value = chessBoard.currentRole === 'self' ? '你的回合' : '对手回合'
    const mh = chessBoard.moveHistoryList || []
    if (mh.length > 0) {
      const last = mh[mh.length - 1]
      lastMove.value = formatMoveLabel(last.from, last.to, last.pieceName, last.pieceColor)
    }
  } catch (e) {
    console.warn('初始化回合/落子显示失败', e)
  }

  // 订阅棋盘事件更新 UI
  ;(channel as any).on('BOARD:ROLE:CHANGE', ({ currentRole }: any) => {
    currentTurn.value = currentRole === 'self' ? '你的回合' : '对手回合'
  })
  ;(channel as any).on('BOARD:MOVE:MADE', ({ from, to, pieceName, pieceColor }: any) => {
    lastMove.value = formatMoveLabel(from, to, pieceName, pieceColor)
  })
  channel.on('NET:GAME:START', ({ color, opponent }) => {
    console.log('Game started, color:', color)
    // 重置游戏状态
    gameOver.value = false
    endModalVisible.value = false
    endResult.value = null
    regretModalVisible.value = false
    drawModalVisible.value = false
    clearModalState()

    chessBoard.stop()
    chessBoard.start(color, true)
    networkPlay.value = true
    opponentInfo.value = opponent
    selfColor.value = color
    // 新增：保存初始游戏状态
    saveGameState({
      isNetPlay: chessBoard.isNetworkPlay(),
      roomId: ws.getCurrentRoomId(), // 需确保ws实例有currentRoomId属性（后续步骤补充）
      selfColor: color,
      moveHistory: chessBoard.moveHistoryList, // 初始为空
      currentRole: chessBoard.currentRole,
      opponentInfo: opponent,
    })
    // 清空消息队列
    channel.clearQueue('NET:GAME:END')
    currentTurn.value = chessBoard.currentRole === 'self' ? '你的回合' : '对手回合'
    lastMove.value = '无'
  })

  // 处理服务端同步消息（重连时）
  channel.on('NET:GAME:SYNC', (data: any) => {
    const { role, currentTurn, roomId } = data
    // 如果服务端的 roomId 与当前客户端的 roomId 不匹配，忽略该 SYNC（可能是上局残留）
    const currentRoom = ws.getCurrentRoomId && ws.getCurrentRoomId()
    if (roomId !== undefined && currentRoom !== undefined && roomId !== currentRoom) {
      console.log('Ignored NET:GAME:SYNC for different room', roomId, currentRoom)
      return
    }
    // 如果本地保存的状态明确表示这是本地对局，则忽略服务端的 SYNC，避免错误切换到联机模式
    const localSaved = getGameState()
    if (localSaved && localSaved.isNetPlay === false) {
      console.log('Ignoring NET:GAME:SYNC because local saved state indicates local play')
      return
    }
    // 以服务端为准，重新启动网络棋盘并尝试恢复本地保存的棋谱
    chessBoard.stop()
    chessBoard.start(role || 'red', true)
    networkPlay.value = true
    selfColor.value = role || 'red'
    // 尝试从 sessionStorage 恢复历史（若存在）
    const savedState = getGameState()
    if (savedState && savedState.isNetPlay) {
      chessBoard.restoreState(savedState)
    }
    // 根据服务端的 currentTurn 设置当前执行方
    if (currentTurn && role) {
      if (currentTurn === role) {
        chessBoard.setCurrentRole('self')
      }
      else {
        chessBoard.setCurrentRole('enemy')
      }
      const isMyTurn = currentTurn === role
      currentTurn.value = isMyTurn ? '你的回合' : '对手回合'
    } else {
      currentTurn.value = chessBoard.currentRole === 'self' ? '你的回合' : '对手回合'
    }
  })

  // 监听聊天消息
  channel.on('NET:CHAT:MESSAGE', ({ sender, content }) => {
    if (networkPlay.value && !gameOver.value) {
      chatPanelRef.value?.receiveMessage(sender, content)
    }
  })
  // 监听和棋请求
  channel.on('NET:DRAW:REQUEST', () => {
    if (gameOver.value)
      return
    // 展示与悔棋一致的模态交互
    drawModalType.value = 'responding'
    drawModalVisible.value = true
    showMsg('对方请求和棋')
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
  })

  channel.on('NET:DRAW:RESPONSE', (data: any) => {
    // 如果自己是请求方，收到响应时关闭请求模态
    drawModalVisible.value = false
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
    if (data.accepted) {
      showMsg('对方同意和棋，局面以和棋结束')
      gameOver.value = true
      chessBoard?.disableInteraction()
    }
    else {
      showMsg('对方拒绝了和棋请求')
    }
  })
  channel.on('NET:GAME:END', ({ winner }) => {
    if (chessBoard.isNetworkPlay() === false) {
      return
    }
    gameOver.value = true
    if (winner === 'red') {
      showMsg('红方胜利')
    }
    else if (winner === 'black') {
      showMsg('黑方胜利')
    }
    else {
      showMsg('和棋')
    }
    chessBoard?.disableInteraction()
    // 保存游戏结束前的当前历史到 sessionStorage，以便复盘页面读取
    saveGameState({
      isNetPlay: chessBoard.isNetworkPlay(),
      selfColor: chessBoard.SelfColor,
      moveHistory: chessBoard.moveHistoryList,
      currentRole: chessBoard.currentRole,
      opponentInfo: opponentInfo.value,
    })
    // 计算当前客户端对局结果
    if (winner === 'draw') {
      endResult.value = 'draw'
    }
    else {
      const myColor = chessBoard.SelfColor
      endResult.value = winner === myColor ? 'win' : 'lose'
    }
    endModalVisible.value = true
    // 持久化结束模态，刷新后仍显示
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
  })

  channel.on('LOCAL:GAME:END', ({ winner }: any) => {
    if (chessBoard.isNetworkPlay()) {
      return
    }
    gameOver.value = true
    if (winner === 'red') {
      showMsg('红方胜利')
    }
    else if (winner === 'black') {
      showMsg('黑方胜利')
    }
    else {
      showMsg('和棋')
    }
    chessBoard?.disableInteraction()
  })
  // 监听悔棋请求
  channel.on('NET:CHESS:REGRET:REQUEST', () => {
    regretModalType.value = 'responding'
    regretModalVisible.value = true
    showMsg('对方请求悔棋')
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
  })
  // 监听悔棋响应
  channel.on('NET:CHESS:REGRET:RESPONSE', (data) => {
    regretModalVisible.value = false
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value, endModalVisible: endModalVisible.value, endResult: endResult.value })
    if (data.accepted) {
      // 对方同意悔棋，执行悔棋操作
      const steps = chessBoard.isMyTurn() ? 2 : 1
      for (let i = 0; i < steps; i++) {
        chessBoard?.regretMove()
      }
      showMsg(`悔了${steps}步棋`)
      chessBoard?.setCurrentRole('self')
      // showMsg(`对方同意悔棋${chessBoard.currentRole}`)
    }
    else {
      showMsg('对方拒绝悔棋')
    }
  })
  window.history.pushState(null, '', window.location.href)
  // 监听 popstate 事件，防止后退操作
  window.addEventListener('popstate', (event) => {
    handlePopState(event)
  })
})
onUnmounted(() => {
  channel.off('NET:GAME:START')
  channel.off('NET:GAME:END')
  channel.off('LOCAL:GAME:END')
  channel.off('NET:CHESS:REGRET:REQUEST')
  channel.off('NET:CHESS:REGRET:RESPONSE')
  channel.off('NET:DRAW:REQUEST')
  channel.off('NET:DRAW:RESPONSE')
  ;(channel as any).off('BOARD:ROLE:CHANGE')
  ;(channel as any).off('BOARD:MOVE:MADE')
  window.removeEventListener('popstate', handlePopState)
  chessBoard?.stop()
})
</script>

<template>
  <div class="h-full w-full flex flex-col sm:flex-row">
    <div class="block h-1/5 sm:h-full flex-1" />
    <div class="relative h-3/5 w-full sm:h-full sm:w-5/12 flex flex-col">
      <div class="relative flex-1 w-full">
        <canvas
          ref="background"
          class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2"
        />
        <canvas ref="chesses" class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2" />
      </div>
      <div class="flex justify-center space-x-4 mb-20">
        <!-- 新增悔棋按钮 -->
        <button
          class="border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="regret"
        >
          悔棋
        </button>
        <!-- 原有的认输按钮 -->
        <button
          v-if="networkPlay"
          class="border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="confirmGiveUp"
        >
          认输
        </button>
        <!-- 新增和棋按钮 -->
        <button
          v-if="networkPlay"
          class="border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="offerDraw"
        >
          和棋
        </button>
        <!-- 原有的退出按钮 -->
        <button
          class="border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="quit"
        >
          退出
        </button>
      </div>
    </div>
    <div class="sm:h-full flex-1 flex flex-col pt-12 pb-20 pr-48">
      <!-- 新增：用户信息面板 -->
      <div v-if="networkPlay" class="bg-white/80 backdrop-blur rounded-xl shadow-sm p-4 mb-4 flex flex-col border border-gray-200">
        <div class="flex items-center justify-between w-full mb-4">
          <div class="flex flex-col items-center w-1/3">
            <img :src="userStore.userInfo?.avatar || '/images/default_avatar.png'" class="w-12 h-12 rounded-full mb-1 object-cover border-2 border-red-500" />
            <span class="text-xs truncate w-full text-center font-medium">{{ userStore.userInfo?.name }}</span>
            <span class="text-xs font-bold mt-1" :class="selfColor === 'red' ? 'text-red-600' : 'text-black'">{{ selfColor === 'red' ? '红方' : '黑方' }}</span>
          </div>
          <div class="text-2xl font-black text-gray-400 italic mx-2">VS</div>
          <div class="flex flex-col items-center w-1/3">
            <img :src="opponentInfo?.avatar || '/images/default_avatar.png'" class="w-12 h-12 rounded-full mb-1 object-cover border-2 border-black" />
            <span class="text-xs truncate w-full text-center font-medium">{{ opponentInfo?.name || '对手' }}</span>
            <span class="text-xs font-bold mt-1" :class="selfColor === 'red' ? 'text-black' : 'text-red-600'">{{ selfColor === 'red' ? '黑方' : '红方' }}</span>
          </div>
        </div>
          <div class="flex justify-between w-full space-x-2">
          <div class="w-1/2 text-xs space-y-1 bg-gray-50 p-2 rounded-lg">
            <div class="flex justify-between items-center">
              <span class="text-gray-500">经验</span>
              <span class="font-bold text-gray-700">{{ userStore.userInfo?.exp || 0 }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">场次</span>
              <span class="font-bold text-gray-700">{{ userStore.userInfo?.totalGames || 0 }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">胜率</span>
              <span class="font-bold text-gray-700">{{ (userStore.userInfo?.winRate || 0).toFixed(1) }}%</span>
            </div>
          </div>
          <div class="w-1/2 text-xs space-y-1 bg-gray-50 p-2 rounded-lg">
            <div class="flex justify-between items-center">
              <span class="text-gray-500">经验</span>
              <span class="font-bold text-gray-700">{{ opponentInfo?.exp || 0 }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">场次</span>
              <span class="font-bold text-gray-700">{{ opponentInfo?.totalGames || 0 }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">胜率</span>
              <span class="font-bold text-gray-700">{{ (opponentInfo?.winRate || 0).toFixed(1) }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 游戏信息（当前回合 / 最近落子） - 仅联机模式显示 -->
      <div v-if="networkPlay" class="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
        <div class="text-sm font-semibold text-blue-900 mb-2">游戏信息</div>
        <div class="grid grid-cols-2 gap-2 text-xs">
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

      <!-- 聊天面板 -->
      <div class="flex-1 overflow-hidden flex flex-col">
        <ChatPanel v-if="networkPlay" ref="chatPanelRef" :ws="ws" />
      </div>
    </div>
  </div>
  <RegretModal
    :visible="regretModalVisible"
    :type="regretModalType"
    :on-accept="handleRegretAccept"
    :on-reject="handleRegretReject"
    @close="regretModalVisible = false"
  />
  <RegretModal
    :visible="drawModalVisible"
    :type="drawModalType"
    mode="draw"
    :on-accept="handleDrawAccept"
    :on-reject="handleDrawReject"
    @close="drawModalVisible = false"
  />
  <GameEndModal
    :visible="endModalVisible"
    :result="endResult"
    :on-review="review"
    :on-quit="quit"
    @close="endModalVisible = false"
  />
  <ConfirmModal
    :visible="resignConfirmVisible"
    :title="networkPlay ? '确认认输？' : '你确定要退出吗'"
    :message="networkPlay ? '认输后本局将判负，是否继续？' : ''"
    confirmText="确认"
    cancelText="取消"
    :on-confirm="handleGiveUpConfirm"
    :on-cancel="handleGiveUpCancel"
  />
  <ConfirmModal
    :visible="quitConfirmVisible"
    :title="networkPlay ? '确认退出？' : '你确定要退出吗'"
    :message="networkPlay ? '退出将被判负，对局记录将被保存。是否继续？' : ''"
    confirmText="退出"
    cancelText="取消"
    :on-confirm="handleQuitConfirm"
    :on-cancel="handleQuitCancel"
  />
</template>
