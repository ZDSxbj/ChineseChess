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
import ChatPanel from '@/components/ChatPanel.vue'

const router = useRouter()
const chatPanelRef = ref()

const background = useTemplateRef('background')
const chesses = useTemplateRef('chesses')

const isPC = inject('isPC') as Ref<boolean>

let chessBoard: ChessBoard
const ws = inject('ws') as WebSocketService

const regretModalVisible = ref(false)
const regretModalType = ref<'requesting' | 'responding'>('requesting')

function handleRegretAccept() {
  ws.sendRegretResponse(true)
  const steps = chessBoard.isMyTurn() ? 1 : 2
  for (let i = 0; i < steps; i++) {
    chessBoard?.regretMove()
  }
  showMsg(`悔了${steps}步棋`)
  chessBoard?.setCurrentRole('enemy')
}

function handleRegretReject() {
  ws.sendRegretResponse(false)
}

function handleDrawAccept() {
  ws.sendDrawResponse(true)
  drawModalVisible.value = false
}

function handleDrawReject() {
  ws.sendDrawResponse(false)
  drawModalVisible.value = false
}

function decideSize(isPCBool: boolean) {
  return isPCBool ? 70 : 40
}

watch(isPC, (newIsPC) => {
  const gridSize = decideSize(newIsPC)
  chessBoard?.redraw(gridSize)
})

function giveUp() {
  ws?.giveUp()
}

function quit() {
  clearGameState() // 退出时清除状态
  router.push('/')
}
// 新增悔棋函数
function regret() {
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
    // 新增：保存初始游戏状态
    saveGameState({
      isNetPlay: chessBoard.isNetworkPlay(),
      roomId: ws.getCurrentRoomId(), // 需确保ws实例有currentRoomId属性（后续步骤补充）
      selfColor: color,
      moveHistory: chessBoard.moveHistoryList, // 初始为空
      currentRole: chessBoard.currentRole,
    })
  })

    }
  })

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
    </div>
  </div>
  <RegretModal
    :visible="regretModalVisible"
    :type="regretModalType"
    :on-accept="handleRegretAccept"
    :on-reject="handleRegretReject"
    @close="regretModalVisible = false"
  />
    :visible="drawModalVisible"
    :type="drawModalType"
    :on-accept="handleDrawAccept"
    :on-reject="handleDrawReject"
    @close="drawModalVisible = false"
  />
</template>
