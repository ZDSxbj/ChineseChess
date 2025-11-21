<script lang="ts" setup>
import type { Ref } from 'vue'
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import ChatPanel from '@/components/ChatPanel.vue'
import { showMsg } from '@/components/MessageBox'
import RegretModal from '@/components/RegretModal.vue'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState, getGameState, saveGameState } from '@/store/gameStore'
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
}

function handleRegretReject() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendRegretResponse(false)
  regretModalVisible.value = false
}

function handleDrawAccept() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawResponse(true)
  // 接受和棋，后端会广播 GAME:END
  drawModalVisible.value = false
}

function handleDrawReject() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawResponse(false)
  drawModalVisible.value = false
  showMsg('已拒绝对方的和棋请求')
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
  ws?.giveUp()
}

function offerDraw() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  ws.sendDrawRequest()
  drawModalType.value = 'requesting'
  drawModalVisible.value = true
  showMsg('已发送和棋请求')
}

function quit() {
  if (!gameOver.value) {
    giveUp()
  }
  clearGameState() // 退出时清除状态
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
  channel.on('NET:GAME:START', ({ color }) => {
    console.log('Game started, color:', color)
    chessBoard.stop()
    chessBoard.start(color, true)
    networkPlay.value = true
    // 新增：保存初始游戏状态
    saveGameState({
      isNetPlay: chessBoard.isNetworkPlay(),
      roomId: ws.getCurrentRoomId(), // 需确保ws实例有currentRoomId属性（后续步骤补充）
      selfColor: color,
      moveHistory: chessBoard.moveHistoryList, // 初始为空
      currentRole: chessBoard.currentRole,
    })
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
  })

  channel.on('NET:DRAW:RESPONSE', (data: any) => {
    // 如果自己是请求方，收到响应时关闭请求模态
    drawModalVisible.value = false
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
  })
  // 监听悔棋响应
  channel.on('NET:CHESS:REGRET:RESPONSE', (data) => {
    regretModalVisible.value = false
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
</template>
