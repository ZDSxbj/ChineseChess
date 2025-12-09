<script lang="ts" setup>
import type { Ref } from 'vue'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import GameEndModal from '@/components/GameEndModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { showMsg } from '@/components/MessageBox'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState } from '@/store/gameStore'
import { useUserStore } from '@/store/useStore'
import channel from '@/utils/channel'

declare const window: any

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const background = useTemplateRef('background')
const chesses = useTemplateRef('chesses')

const isPC = inject('isPC') as Ref<boolean>

let chessBoard: ChessBoard

const gameOver = ref(false)
const endModalVisible = ref(false)
const endResult = ref<'win' | 'lose' | 'draw' | null>(null)

// 从路由参数获取难度和颜色
const aiLevel = ref(Number(route.query.difficulty) || 3) // AI难度 1-6，默认3
// 将数值难度映射为用户友好的标签：简单 / 中等 / 困难
const aiLabel = computed(() => {
  const lvl = Number(aiLevel.value) || 3
  if (lvl <= 2) return '简单'
  if (lvl <= 4) return '中等'
  return '困难'
})
const playerColor = ref<'red' | 'black'>((route.query.color as 'red' | 'black') || 'red')
const aiThinking = ref(false) // AI正在思考
const quitConfirmVisible = ref(false)

// 当前回合与最近一步（响应式）
const currentTurn = ref<string>('—')
const lastMove = ref<string>('无')

function formatMoveLabel(from: any, to: any, pieceName?: string, pieceColor?: string) {
  // 简单中文记谱：例如“马二进三”
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
  // 计算从/到 文件编号（1-9），按走子方视角
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

function decideSize(isPCBool: boolean) {
  return isPCBool ? 70 : 40
}

function giveUp() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  // 获胜者是对手
  const opponentColor = playerColor.value === 'red' ? 'black' : 'red'
  channel.emit('LOCAL:GAME:END', { winner: opponentColor, reason: 'resign' })
}

function quit() {
  // 使用自制模态确认退出，确认后直接退出且不计为认输（不触发输的提示）
  if (!gameOver.value) {
    quitConfirmVisible.value = true
    return
  }
  clearGameState()
  router.push('/')
}

function handleQuitConfirm() {
  quitConfirmVisible.value = false
  clearGameState()
  router.push('/')
}

function handleQuitCancel() {
  quitConfirmVisible.value = false
}

function regret() {
  if (gameOver.value) {
    showMsg('游戏已结束')
    return
  }
  if (!chessBoard || chessBoard.stepsNum === 0) {
    showMsg('没有可悔棋的步数')
    return
  }
  // 如果玩家是黑方且只走了一步，不能悔棋（因为要悔掉AI的第一步）
  if (playerColor.value === 'black' && chessBoard.stepsNum === 1) {
    showMsg('没有可悔棋的步数')
    return
  }
  // 使用ChessBoard提供的接口，确保回到玩家上一步之前
  const undone = chessBoard.regretLastTurn()
  if (!undone) {
    showMsg('没有可悔棋的步数')
    return
  }
  showMsg(`悔了${undone}步棋`) 
}

function handlePopState(_event: PopStateEvent) {
  window.history.pushState(null, '', window.location.href)
  showMsg('请通过应用内的导航按钮进行操作')
}

/**
 * 使用logic.js中的AI算法进行走棋
 */
function requestAIMove() {
  console.log('requestAIMove called')
  console.log('window.logic:', window.logic)
  
  if (!window.logic || !window.logic.getAIMoveAdaptive) {
    console.error('AI引擎未加载')
    showMsg('AI引擎加载失败')
    return
  }

  try {
    aiThinking.value = true
    const board = chessBoard.getCurrentBoard()
    const aiColor = playerColor.value === 'red' ? 'black' : 'red'

    console.log('AI thinking for color:', aiColor)
    console.log('Current board:', board)
    console.log('AI level:', aiLevel.value)

    // 使用自适应AI算法（注意：第二个参数是AI所行棋的颜色）
    const moveData = window.logic.getAIMoveAdaptive(board, aiColor, {
      playerSkill: aiLevel.value,
      useOpeningBook: true,
      moveNumber: chessBoard.stepsNum / 2,
    })

    console.log('AI move data:', moveData)

    if (!moveData) {
      console.log('AI无合法移动')
      showMsg('AI无合法移动')
      // 判断是否为将死
      if (window.logic.isCheckmate(board, aiColor)) {
        channel.emit('LOCAL:GAME:END', { winner: playerColor.value })
      }
      return
    }

    console.log('AI moving from', moveData.from, 'to', moveData.to)

    // 尝试将 logic.js 的 [row,col] -> 内部 {x,y} 并应用
    const engineFrom = moveData.from
    const engineTo = moveData.to
    // engine 返回的是 [row, col] -> row 对应前端 y，col 对应前端 x
    let fromX = engineFrom[1]
    let fromY = engineFrom[0]
    let toX = engineTo[1]
    let toY = engineTo[0]
    // 如果玩家执黑，前端与 engine 的纵向坐标需要翻转
    if (chessBoard.SelfColor === 'black') {
      fromY = 9 - fromY
      toY = 9 - toY
    }
    let fromPos = { x: fromX, y: fromY }
    let toPos = { x: toX, y: toY }
    // 先检查前端该坐标是否有棋子
    let pieceAtFrom = chessBoard.getPieceAt(fromPos)
    if (!pieceAtFrom) {
      // 可能坐标方向不同，尝试交换（备用策略）
      const altFrom = { x: moveData.from[0], y: moveData.from[1] }
      const altTo = { x: moveData.to[0], y: moveData.to[1] }
      // 如果玩家执黑，alt 映射也需要翻转 y
      if (chessBoard.SelfColor === 'black') {
        altFrom.y = 9 - altFrom.y
        altTo.y = 9 - altTo.y
      }
      const altPiece = chessBoard.getPieceAt(altFrom)
      if (altPiece) {
        console.warn('Using alternate coordinate mapping for AI move')
        fromPos = altFrom
        toPos = altTo
        pieceAtFrom = altPiece
      }
    }

    let applied = false
    if (pieceAtFrom) {
      applied = chessBoard.applyAIMove(fromPos, toPos)
      if (!applied) {
        console.error('applyAIMove returned false for', fromPos, toPos)
        // 尝试使用 AI 引擎产生的新棋盘作为回退（将直接替换前端棋盘）
        try {
          console.warn('applyAIMove failed, falling back to logic.applyMove + setCurrentBoard')
          const newBoard = window.logic.applyMove(board, moveData.from, moveData.to)
          chessBoard.setCurrentBoard(newBoard)
          chessBoard.render()
          applied = true
        } catch (err) {
          console.error('Fallback setCurrentBoard also failed', err)
          applied = false
        }
      }
    } else {
      // 作为最后手段，直接把 AI 的新棋盘设置到前端（不会产生 moveHistory）
      console.warn('No piece found at AI move from-pos; falling back to setCurrentBoard')
      const newBoard = window.logic.applyMove(board, moveData.from, moveData.to)
        chessBoard.setCurrentBoard(newBoard)
        chessBoard.render()
        // 试图读取落子位置上的棋子名并触发 BOARD:MOVE:MADE，确保最近落子更新
        try {
          const movedPiece = chessBoard.getPieceAt(toPos)
          ;(channel as any).emit('BOARD:MOVE:MADE', { from: fromPos, to: toPos, pieceName: movedPiece?.name, pieceColor: movedPiece?.color })
        } catch (e) {
          console.warn('Emit BOARD:MOVE:MADE after fallback failed', e)
        }
        applied = true
    }

    if (!applied) {
      console.error('Failed to apply AI move to ChessBoard', fromPos, toPos)
      showMsg('AI移动应用失败')
      aiThinking.value = false
      return
    }

    // 更新棋盘显示
    chessBoard.render()

    console.log('AI move applied on frontend from', fromPos, 'to', toPos)

    // 检查玩家是否被将死，使用前端当前棋盘状态
    const newBoard = chessBoard.getCurrentBoard()
    const playerIsCheckmate = window.logic.isCheckmate(newBoard, playerColor.value)
    if (playerIsCheckmate) {
      showMsg('AI赢了')
      const aiColor = playerColor.value === 'red' ? 'black' : 'red'
      channel.emit('LOCAL:GAME:END', { winner: aiColor })
      return
    }

    // 检查玩家是否被将
    const playerInCheck = window.logic.isInCheck(newBoard, playerColor.value)
    if (playerInCheck) {
      showMsg('你被将军了！')
    }

    chessBoard.setCurrentRole('self')
    aiThinking.value = false
  } catch (error) {
    console.error('AI走棋错误:', error)
    showMsg('AI走棋出错')
    aiThinking.value = false
  }
}

onMounted(() => {
  console.log('AIView mounted, route.query:', route.query)
  console.log('playerColor:', playerColor.value, 'aiLevel:', aiLevel.value)
  
  const gridSize = decideSize(isPC.value)
  const canvasBackground = background.value as HTMLCanvasElement
  const canvasChesses = chesses.value as HTMLCanvasElement
  
  console.log('Canvas elements:', canvasBackground, canvasChesses)
  console.log('Grid size:', gridSize)
  
  const ctxBackground = canvasBackground.getContext('2d')
  const ctxChesses = canvasChesses.getContext('2d')

  if (!ctxBackground || !ctxChesses) {
    throw new Error('Failed to get canvas context')
  }

  console.log('Creating ChessBoard...')
  chessBoard = new ChessBoard(canvasBackground, canvasChesses, gridSize)
  console.log('Starting ChessBoard with color:', playerColor.value)
  chessBoard.start(playerColor.value, false, true) // 第三个参数true表示AI模式
  console.log('ChessBoard started successfully')

  window.history.pushState(null, '', window.location.href)

  // 监听本地游戏结束事件
  channel.on('LOCAL:GAME:END', (payload: any) => {
    const { winner, reason } = payload || {}

    // 如果是认输（resign），直接接受结束
    if (reason === 'resign') {
      gameOver.value = true
      if (winner === 'draw') {
        showMsg('和棋')
        endResult.value = 'draw'
      } else if (winner === playerColor.value) {
        showMsg('你赢了!')
        endResult.value = 'win'
      } else {
        showMsg('你输了')
        endResult.value = 'lose'
      }
      chessBoard?.disableInteraction()
      endModalVisible.value = true
      return
    }

    // 否则尽量校验是否真实的将死（避免误触发）
    try {
      const board = chessBoard.getCurrentBoard()
      // 当收到 winner 时，应验证被将死的一方是否真的被将死（loser），而不是验证 winner 本身
      const loser = winner === 'red' ? 'black' : 'red'
      if (winner === 'draw' || (window.logic && window.logic.isCheckmate(board, loser))) {
        gameOver.value = true
        if (winner === 'draw') {
          showMsg('和棋')
          endResult.value = 'draw'
        } else if (winner === playerColor.value) {
          showMsg('你赢了!')
          endResult.value = 'win'
        } else {
          showMsg('你输了')
          endResult.value = 'lose'
        }
        chessBoard?.disableInteraction()
        endModalVisible.value = true
      } else {
        console.warn('Ignored LOCAL:GAME:END because board is not in checkmate for loser', loser, 'payload:', payload)
      }
    } catch (e) {
      console.error('Error validating LOCAL:GAME:END payload', payload, e)
      // 若校验失败则保守处理：接受结束，避免用户体验异常（可改为忽略）
      gameOver.value = true
      endResult.value = winner === playerColor.value ? 'win' : 'lose'
      chessBoard?.disableInteraction()
      endModalVisible.value = true
    }
  })

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


  // 监听玩家移动完成后的AI思考
  channel.on('GAME:MOVE', () => {
    console.log('GAME:MOVE event received')
    
    // 检查logic是否已加载
    if (!window.logic || !window.logic.isCheckmate) {
      console.error('window.logic not loaded in GAME:MOVE handler')
      showMsg('AI引擎未加载')
      return
    }
    
    // 检查玩家是否被将死
    const board = chessBoard.getCurrentBoard()
    const playerIsCheckmate = window.logic.isCheckmate(board, playerColor.value)

    if (playerIsCheckmate) {
      showMsg('AI赢了')
      const aiColor = playerColor.value === 'red' ? 'black' : 'red'
      channel.emit('GAME:END', { winner: aiColor, online: false })
      return
    }

    console.log('Switching to AI turn')
    // 切换到AI思考
    chessBoard.setCurrentRole('enemy')
    setTimeout(() => {
      requestAIMove()
    }, 500)
  })

  // 如果玩家是黑方，由AI先手
  if (playerColor.value === 'black') {
    setTimeout(() => {
      chessBoard.setCurrentRole('enemy')
      requestAIMove()
    }, 1000)
  }

  window.addEventListener('popstate', handlePopState)
})

onUnmounted(() => {
  channel.off('LOCAL:GAME:END')
  channel.off('GAME:MOVE')
  ;(channel as any).off('BOARD:ROLE:CHANGE')
  ;(channel as any).off('BOARD:MOVE:MADE')
  window.removeEventListener('popstate', handlePopState)
  chessBoard?.stop()
})
// 模态组件放在模板最后
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
          class="border-0 rounded-2xl bg-red-500 text-white p-4 transition-all duration-200"
          text="xl"
          hover="bg-red-600"
          @click="giveUp"
        >
          认输
        </button>
        <button
          class="border-0 rounded-2xl bg-gray-500 text-white p-4 transition-all duration-200"
          text="xl"
          hover="bg-gray-600"
          @click="quit"
        >
          退出
        </button>
      </div>
    </div>
    <div class="sm:h-full flex-1 flex flex-col pt-12 pb-20 pr-48">
      <!-- AI对战信息面板 -->
      <div class="bg-white/80 backdrop-blur rounded-xl shadow-sm p-4 mb-4 flex flex-col border border-gray-200">
        <div class="flex items-center justify-between w-full mb-4">
          <div class="flex flex-col items-center w-1/3">
            <img :src="userStore.userInfo?.avatar || '/images/default_avatar.png'" alt="玩家头像" class="w-12 h-12 rounded-full mb-1 object-cover border-2 border-red-500" />
            <span class="text-xs truncate w-full text-center font-medium">{{ userStore.userInfo?.name }}</span>
            <span class="text-xs font-bold mt-1" :class="playerColor === 'red' ? 'text-red-600' : 'text-black'">{{ playerColor === 'red' ? '红方' : '黑方' }}</span>
          </div>
          <div class="text-2xl font-black text-gray-400 italic mx-2">VS</div>
          <div class="flex flex-col items-center w-1/3">
            <div class="w-12 h-12 rounded-full mb-1 bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center text-white font-bold text-lg">
              🤖
            </div>
            <span class="text-xs truncate w-full text-center font-medium">电脑</span>
            <span class="text-xs font-bold mt-1" :class="playerColor === 'red' ? 'text-black' : 'text-red-600'">{{ playerColor === 'red' ? '黑方' : '红方' }}</span>
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
              <span class="font-bold text-gray-700">{{ (userStore.userInfo?.winRate || 0).toFixed(1) }}%</span>
            </div>
          </div>
          <div class="w-1/2 text-xs space-y-1 bg-gray-50 p-2 rounded-lg">
            <div class="flex justify-between items-center">
              <span class="text-gray-500">AI难度</span>
              <span class="font-bold text-gray-700">{{ aiLabel }}</span>
            </div>
            <div class="flex justify-between items-center">
              <span class="text-gray-500">棋手</span>
              <span class="font-bold text-gray-700">超级大脑</span>
            </div>
            <!-- 电脑胜率已隐藏（由设计要求移除） -->
          </div>
        </div>
      </div>

      <!-- 游戏统计 -->
      <div class="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
        <div class="text-sm font-semibold text-blue-900 mb-2">游戏信息</div>
        <div class="grid grid-cols-2 gap-2 text-xs">
          <div class="flex justify-between">
            <span class="text-gray-600">已走步数:</span>
            <span class="font-bold">{{ chessBoard?.stepsNum || 0 }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">游戏模式:</span>
            <span class="font-bold">人机对战</span>
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
    :visible="endModalVisible"
    :result="endResult"
    :on-review="() => router.push('/game/replay')"
    :on-quit="quit"
    @close="endModalVisible = false"
  />
  <ConfirmModal
    :visible="quitConfirmVisible"
    title="确定要退出当前对局？"
    message="退出后将返回首页，比赛结果不会被记录。"
    confirmText="退出"
    cancelText="取消"
    :on-confirm="handleQuitConfirm"
    :on-cancel="handleQuitCancel"
  />
</template>
