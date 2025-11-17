<script lang="ts" setup>
import type { Ref } from 'vue'
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import channel from '@/utils/channel'
import { showMsg } from '@/components/MessageBox'
import DrawModal from '@/components/DrawModal.vue'
import RegretModal from '@/components/RegretModal.vue'

import ChessBoard from '@/composables/ChessBoard'

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

const drawModalVisible = ref(false) // 新增：和棋弹窗显示状态
const drawModalType = ref<'requesting' | 'responding'>('requesting') // 新增：和棋弹窗类型
const gameEnded = ref(false) // 新增：游戏结束状态

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

// 新增：处理和棋同意
function handleDrawAccept() {
  // 同意和棋，通知对方
  ws.sendDrawResponse(true)
  // 游戏结束，显示和棋消息
  showMsg('双方同意和棋，游戏结束！')
  gameEnded.value = true // 设置游戏结束状态
  // 新增：禁用棋盘操作
  chessBoard?.endGame('draw')
  drawModalVisible.value = false
}

// 新增：处理和棋拒绝
function handleDrawReject() {
  // 拒绝和棋，通知对方
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
  if (gameEnded.value) return // 游戏结束后禁用
  ws?.giveUp()
  // 新增：处理认输后的游戏结束状态
  const loserColor = chessBoard?.Color; // 自己的颜色（认输方）
  const winnerColor = loserColor === 'red' ? 'black' : 'red'; // 对方颜色（胜方）
  // 触发本地GAME:END事件，明确是认输
  channel.emit('GAME:END', {
    winner: winnerColor,
    isResign: true
  });
}

function quit() {
  giveUp()
  clearGameState() // 退出时清除状态
  const currentState = history.state
  window.history.pushState(currentState, '', window.location.href)
  router.push('/')
}
// 新增悔棋函数
function regret() {
  if (gameEnded.value) return // 游戏结束后禁用
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

// 处理回退事件
function handlePopState(event: PopStateEvent) {
  // 阻止回退
  const currentState = history.state
  window.history.pushState(currentState, '', window.location.href)
  // 提示用户，防止意外退出
  showMsg('请通过游戏内的退出按钮退出游戏')
}
function handleBeforeUnload(event: BeforeUnloadEvent) {
  event.preventDefault()
  // event.returnValue = '不要刷新' // 标准方式
  return '不要刷新' // 兼容某些浏览器
}
// 新增：和棋函数
function draw() {
  if (gameEnded.value) return // 游戏结束后禁用

  if (chessBoard) {
    if (chessBoard.isNetworkPlay()) {
      // 联网模式：发送和棋请求给对手
      ws.sendDrawRequest()
      // 显示等待提示
      drawModalType.value = 'requesting'
      drawModalVisible.value = true
    }
    else {
      // 本地模式：直接显示和棋消息
      showMsg('本地模式无法和棋')
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

  // 监听游戏结束事件（将死/认输时触发）
channel.on('GAME:END', ({ winner, isResign = false }) => {
  gameEnded.value = true;
  const result = winner === chessBoard?.Color ? 'win' : 'lose';
  chessBoard?.endGame(result);

  // 根据是否为认输，显示不同消息
  if (isResign) {
    // 认输场景：明确提示“对方认输”
    if (result === 'win') {
      showMsg('对方已认输，你胜利了！');
    } else {
      showMsg('你已认输，游戏结束！');
    }
  } else {
    // 将死场景：提示“胜利”
    showMsg(`${winner === 'red' ? '红' : '黑'}方胜利，游戏结束！`);
  }
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
  window.history.pushState(null, '', window.location.href) // 修改浏览器历史记录
  window.addEventListener('popstate', handlePopState)
  // 在页面刷新时给出提示
  window.addEventListener('beforeunload', handleBeforeUnload)
  // 监听和棋请求
  channel.on('NET:CHESS:DRAW:REQUEST', () => {
    drawModalType.value = 'responding'
    drawModalVisible.value = true
    showMsg('对方请求和棋')
  })
  // 监听和棋响应
  channel.on('NET:CHESS:DRAW:RESPONSE', (data) => {
    drawModalVisible.value = false
    if (data.accepted) {
      // 对方同意和棋，游戏结束
      showMsg('双方同意和棋，游戏结束！')
      gameEnded.value = true // 设置游戏结束状态
      // 新增：禁用棋盘操作
      chessBoard?.endGame('draw')
    }
    else {
      showMsg('对方拒绝和棋')
    }
  })
})

onUnmounted(() => {
  window.removeEventListener('popstate', handlePopState)
  window.removeEventListener('beforeunload', handleBeforeUnload)
  channel.off('NET:GAME:START')
  channel.off('GAME:END')
  channel.off('NET:GAME:END')
  channel.off('NET:CHAT:MESSAGE')
  channel.off('NET:CHESS:REGRET:REQUEST')
  channel.off('NET:CHESS:REGRET:RESPONSE')
  channel.off('NET:CHESS:DRAW:REQUEST') // 新增：取消和棋请求监听
  channel.off('NET:CHESS:DRAW:RESPONSE') // 新增：取消和棋响应监听
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
    <div class="sm:h-full sm:w-1/5">
      <!-- 新增悔棋按钮 -->
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        :class="gameEnded ? 'bg-gray-400 cursor-not-allowed' : 'bg-gray-2 hover:bg-gray-9 hover:text-gray-2'"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        @click="regret"
        :disabled="gameEnded"
      >
        悔棋
      </button>
      <!-- 和棋按钮 -->
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        :class="gameEnded ? 'bg-gray-400 cursor-not-allowed' : 'bg-gray-2 hover:bg-gray-9 hover:text-gray-2'"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        @click="draw"
        :disabled="gameEnded"
      >
        和棋
      </button>
      <!-- 认输按钮 -->
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        :class="gameEnded ? 'bg-gray-400 cursor-not-allowed' : 'bg-gray-2 hover:bg-gray-9 hover:text-gray-2'"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        @click="giveUp"
        :disabled="gameEnded"
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
  <RegretModal
    :visible="regretModalVisible"
    :type="regretModalType"
    :on-accept="handleRegretAccept"
    :on-reject="handleRegretReject"
    @close="regretModalVisible = false"
  />
  <DrawModal
    :visible="drawModalVisible"
    :type="drawModalType"
    :on-accept="handleDrawAccept"
    :on-reject="handleDrawReject"
    @close="drawModalVisible = false"
  />
</template>
