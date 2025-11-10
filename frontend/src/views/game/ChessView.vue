<script lang="ts" setup>
import type { Ref } from 'vue'
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showMsg } from '@/components/MessageBox'
import RegretModal from '@/components/RegretModal.vue'
import ChessBoard from '@/composables/ChessBoard'
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
  // 同意悔棋，通知对方
  ws.sendRegretResponse(true)
  // 自己也执行悔棋
  const steps = chessBoard.isMyTurn() ? 1 : 2
  for (let i = 0; i < steps; i++) {
    chessBoard?.regretMove()
  }
  showMsg(`悔了${steps}步棋`)
  // 悔棋后固定为对方的回合
  chessBoard?.setCurrentRole('enemy')
  regretModalVisible.value = false // 发送响应后销毁提示框
}

function handleRegretReject() {
  // 拒绝悔棋，通知对方
  ws.sendRegretResponse(false)
  regretModalVisible.value = false // 发送响应后销毁提示框
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
  giveUp()
  router.push('/')
}
// 新增悔棋函数
function regret() {
  if (chessBoard) {
    if (!chessBoard.stepsNum) {
      showMsg('没有可悔棋的步数')
    }
    else if (chessBoard.isNetworkPlay()) {
      // 新增判断：联网状态下，自身是黑子且历史步数为1时不能悔棋
      if (chessBoard.Color === 'black' && chessBoard.stepsNum === 1) {
        showMsg('没有可悔棋的步数')
        return
      }
      // 联网模式：发送悔棋请求给对手
      ws.sendRegretRequest()
      // 显示等待提示
      regretModalType.value = 'requesting'
      regretModalVisible.value = true
    }
    else {
      // 本地模式：直接执行悔棋
      chessBoard.regretMove()
    }
  }
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
  channel.on('NET:GAME:START', ({ color }) => {
    console.log('Game started, color:', color)
    chessBoard.stop()
    chessBoard.start(color, true)
  })

  // 监听聊天消息
  channel.on('NET:CHAT:MESSAGE', ({ sender, content }) => {
    chatPanelRef.value?.receiveMessage(sender, content)
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
})
onUnmounted(() => {
  channel.off('NET:GAME:START')
  channel.off('NET:CHESS:REGRET:REQUEST')
  channel.off('NET:CHESS:REGRET:RESPONSE')
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
          class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
          text="black xl"
          hover="bg-gray-9 text-gray-2"
          @click="giveUp"
        >
          认输
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
        <ChatPanel ref="chatPanelRef" :ws="ws" />
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
</template>
