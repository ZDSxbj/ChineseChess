<script lang="ts" setup>
import type { Ref } from 'vue'
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import ChatPanel from '@/components/ChatPanel.vue'
import GameEndModal from '@/components/GameEndModal.vue'
import { showMsg } from '@/components/MessageBox'
import RegretModal from '@/components/RegretModal.vue'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState, getGameState, saveGameState, saveModalState, getModalState, clearModalState } from '@/store/gameStore'
import channel from '@/utils/channel'

const router = useRouter()
const chatPanelRef = ref()
const networkPlay = ref(false)

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
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
}

function handleRegretReject() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendRegretResponse(false)
  regretModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
}

function handleDrawAccept() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawResponse(true)
  // 接受和棋，后端会广播 GAME:END
  drawModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
}

function handleDrawReject() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawResponse(false)
  drawModalVisible.value = false
  saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
  showMsg('已拒绝对方的和棋请求')
}

function review() {
  // 确保保存最新的局面到 sessionStorage
  saveGameState({
    isNetPlay: chessBoard.isNetworkPlay(),
    selfColor: chessBoard.SelfColor,
    moveHistory: chessBoard.moveHistoryList,
    currentRole: chessBoard.currentRole,
  })
  router.push('/game/replay')
}

function decideSize(isPCBool: boolean) {
  return isPCBool ? 70 : 40
}

watch(isPC, (newIsPC) => {
  const gridSize = decideSize(newIsPC)
  chessBoard?.redraw(gridSize)
})

function giveUp() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  if (networkPlay.value) {
    ws?.giveUp()
  }
  else {
    // 本地对局直接触发本地结束事件，胜者为对手
    const opponentColor = chessBoard.SelfColor === 'red' ? 'black' : 'red'
    channel.emit('GAME:END', { winner: opponentColor, online: false })
  }
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
    giveUp()
  }
  clearGameState() // 退出时清除状态
  clearModalState() // 退出时清除模态状态，防止下次进入时残留
  // 关键：临时注册空处理，消费队列中缓存的 NET:GAME:END 事件
  // const emptyHandler = () => {
  //   showMsg('捕捉到了信息')
  // }
  // channel.on('NET:GAME:END', emptyHandler)
  // // 立即移除空处理，避免影响后续逻辑
  // channel.off('NET:GAME:END') // 注意：需要 channel 支持移除指定 handler
  // 关键修复：清空NET:GAME:END事件队列，避免残留
  channel.clearQueue('NET:GAME:END')
  router.push('/')
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
      saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
    }
    else {
      chessBoard.regretMove()
    }
  }
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
    console.log('Game state restored from localStorage')
  }

  // 恢复弹窗状态（若有）
  const modalState = getModalState()
  if (modalState) {
    regretModalVisible.value = !!modalState.regretModalVisible
    regretModalType.value = modalState.regretModalType || 'requesting'
    drawModalVisible.value = !!modalState.drawModalVisible
    drawModalType.value = modalState.drawModalType || 'requesting'
  }

  // 新增：保存初始游戏状态
  saveGameState({
    isNetPlay: chessBoard.isNetworkPlay(),
    selfColor: chessBoard.SelfColor,
    moveHistory: chessBoard.moveHistoryList, // 初始为空
    currentRole: chessBoard.currentRole,
  })
  // 清空消息队列
  channel.clearQueue('NET:GAME:END')
  channel.on('NET:GAME:START', ({ color }) => {
    console.log('Game started, color:', color)
    chessBoard.stop()
    chessBoard.start(color, true)
    networkPlay.value = true
    // 重置结束模态与结果，避免上局结束弹窗残留
    endModalVisible.value = false
    endResult.value = null
    gameOver.value = false
    // 新增：保存初始游戏状态
    saveGameState({
      isNetPlay: chessBoard.isNetworkPlay(),
      roomId: ws.getCurrentRoomId(), // 需确保ws实例有currentRoomId属性（后续步骤补充）
      selfColor: color,
      moveHistory: chessBoard.moveHistoryList, // 初始为空
      currentRole: chessBoard.currentRole,
    })
    // 清空消息队列
    channel.clearQueue('NET:GAME:END')
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
    // 重置上局的结束弹窗与状态，避免残留影响新的联机对局
    endModalVisible.value = false
    endResult.value = null
    gameOver.value = false
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
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
  })

  channel.on('NET:DRAW:RESPONSE', (data: any) => {
    // 如果自己是请求方，收到响应时关闭请求模态
    drawModalVisible.value = false
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
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
    // 游戏结束时清理模态状态，避免残留在 sessionStorage 中
    clearModalState()
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
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
  })
  // 监听悔棋响应
  channel.on('NET:CHESS:REGRET:RESPONSE', (data) => {
    regretModalVisible.value = false
    saveModalState({ regretModalVisible: regretModalVisible.value, regretModalType: regretModalType.value, drawModalVisible: drawModalVisible.value, drawModalType: drawModalType.value })
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
  window.removeEventListener('popstate', handlePopState)
  chessBoard?.stop()
})
</script>

<template>
  <div class="h-full w-full flex flex-col sm:flex-row">
    <div class="block h-1/5 sm:h-full sm:w-1/5" />
    <div class="relative h-3/5 w-full sm:h-full sm:w-3/5">
      <canvas
        ref="background"
        class="absolute left-1/2 top-1/4 -translate-x-1/2 -translate-y-1/4"
      />
      <canvas ref="chesses" class="absolute left-1/2 top-1/4 -translate-x-1/2 -translate-y-1/4" />
    </div>
    <div class="sm:h-full sm:w-1/5 flex flex-col">
      <div class="flex flex-col space-y-4 mb-4">
        <!-- 新增悔棋按钮 -->
        <button
          class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="regret"
        >
          悔棋
        </button>
        <!-- 原有的认输按钮 -->
        <button
          v-if="networkPlay"
          class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="giveUp"
        >
          认输
        </button>
        <!-- 新增和棋按钮 -->
        <button
          v-if="networkPlay"
          class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="offerDraw"
        >
          和棋
        </button>
        <!-- 原有的退出按钮 -->
        <button
          class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="quit"
        >
          退出
        </button>
      </div>

      <!-- 聊天面板 -->
      <div class="flex-1">
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
</template>
