<script lang="ts" setup>
import type { Ref } from 'vue'
import { inject, onMounted, onUnmounted, ref, useTemplateRef, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import GameEndModal from '@/components/GameEndModal.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { showMsg } from '@/components/MessageBox'
import ChessBoard from '@/composables/ChessBoard'
import { clearGameState, saveGameState } from '@/store/gameStore'
import type { GameState } from '@/store/gameStore'
import { useUserStore } from '@/store/useStore'
import channel from '@/utils/channel'
import { saveGameRecord } from '@/api/user/getGameRecords'
import { getProfile } from '@/api/user/getProfile'

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
const resignConfirmVisible = ref(false)

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

// 维护行棋历史（格式：紧凑字符串，如 "6665"）
const moveHistory = ref<string>('')
const gameStartTime = ref<Date>(new Date())
let recordSaved = false // 标记是否已保存，防止重复保存
let isReplaying = false // 标记是否正在恢复棋谱

// 【问题1修复】维护有效行棋历史数组（记录每一步的完整信息）
// 用于准确计算步数，避免悔棋时历史混乱
const validMoveHistory = ref<Array<{from: any, to: any, pieceName: string, pieceColor: string}>>([])

// 当前回合与最近一步（响应式）
const currentTurn = ref<string>('—')
const lastMove = ref<string>('无')
const moveCount = ref(0)

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
  resignConfirmVisible.value = true
}

function handleResignConfirm() {
  resignConfirmVisible.value = false
  // 获胜者是对手
  const opponentColor = playerColor.value === 'red' ? 'black' : 'red'
  channel.emit('LOCAL:GAME:END', { winner: opponentColor, reason: 'resign' })
}

function handleResignCancel() {
  resignConfirmVisible.value = false
}

function quit() {
  // 使用自制模态确认退出，确认后直接退出且不计为认输（不触发输的提示）
  if (!gameOver.value) {
    quitConfirmVisible.value = true
    return
  }
  // 【问题2修复】对局结束时清除缓存
  clearAIGameStateFromSession()
  clearGameState()
  router.push('/')
}

function handleQuitConfirm() {
  quitConfirmVisible.value = false
  // 【问题2修复】主动退出时清除缓存
  clearAIGameStateFromSession()
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

  // 【问题1修复】悔棋时从有效历史中删除对应步数
  // 玩家悔棋通常撤销2步（AI的一步 + 玩家的一步），或1步（仅玩家的一步）
  const stepsToPop = undone
  for (let i = 0; i < stepsToPop; i++) {
    validMoveHistory.value.pop()
  }

  // 同步删除紧凑字符串历史，确保存储中的历史与实际局面一致（每步4位坐标）
  const charsToTrim = undone * 4
  if (charsToTrim > 0 && moveHistory.value.length >= charsToTrim) {
    moveHistory.value = moveHistory.value.slice(0, moveHistory.value.length - charsToTrim)
  }

  showMsg(`悔了${undone}步棋`)

  // 【新增】悔棋后立即保存当前对局状态，确保刷新后能恢复悔棋后的状态
  // 仿照联机对战的 saveModalState，这里调用 saveAIGameStateToSession
  saveAIGameStateToSession()
}

function handlePopState(_event: PopStateEvent) {
  window.history.pushState(null, '', window.location.href)
  showMsg('请通过应用内的导航按钮进行操作')
}

/**
 * 【复盘修复】保存游戏状态到 gameStore，供复盘功能使用
 * 在游戏结束时调用，确保复盘功能可以读取完整的行棋历史（包括最后一步绝杀棋）
 */
function saveGameStateForReplay() {
  try {
    // 将 validMoveHistory 转换为 gameStore 所需的格式
    const moveHistoryForReplay = validMoveHistory.value.map(move => ({
      from: move.from,
      to: move.to,
      capturedPiece: null, // AI 对战不需要记录被吃的棋子详情
      currentRole: 'self' as const, // 对于复盘来说，角色信息不重要
    }))

    const gameState: GameState = {
      isNetPlay: false,
      selfColor: playerColor.value,
      moveHistory: moveHistoryForReplay,
      currentRole: 'self',
    }

    saveGameState(gameState)
    console.log('游戏状态已保存到 gameStore，用于复盘', { moveCount: validMoveHistory.value.length })
  } catch (e) {
    console.warn('保存游戏状态到 gameStore 失败:', e)
  }
}

/**
 * 【问题2修复】保存当前人机对战状态到 sessionStorage
 * 这样刷新页面时可以恢复对局
 */
function saveAIGameStateToSession() {
  const serializedMoves = chessBoard?.moveHistoryList
    ? chessBoard.moveHistoryList.map((m: any) => ({
        from: m.from,
        to: m.to,
        currentRole: m.currentRole,
        pieceName: m.pieceName,
        pieceColor: m.pieceColor,
      }))
    : []

  const state = {
    playerColor: playerColor.value,
    aiLevel: aiLevel.value,
    gameStartTime: gameStartTime.value.toISOString(),
    moveHistory: moveHistory.value,
    validMoveHistory: validMoveHistory.value,
    currentTurn: currentTurn.value,
    lastMove: lastMove.value,
    gameOver: gameOver.value,
    // 存储棋盘状态用于恢复
    boardState: chessBoard ? {
      stepsNum: chessBoard.stepsNum,
      currentRole: chessBoard.currentRole,
      moveHistoryList: serializedMoves,
    } : null,
  }
  try {
    sessionStorage.setItem('aiGameState', JSON.stringify(state))
    console.log('人机对战状态已保存到 sessionStorage')
  } catch (e) {
    console.warn('保存人机对战状态失败:', e)
  }
}

/**
 * 【问题2修复】清除 sessionStorage 中的对局状态
 * 在对局结束或主动退出时调用，避免缓存污染
 */
function clearAIGameStateFromSession() {
  try {
    sessionStorage.removeItem('aiGameState')
    console.log('人机对战状态已清除')
  } catch (e) {
    console.warn('清除人机对战状态失败:', e)
  }
}

/**
 * 保存人机对战记录到后端
 * 【问题1修复】使用 validMoveHistory 来计算最终有效步数
 * 悔棋时 validMoveHistory 已正确删除，保证保存的历史是准确的
 */
async function saveAIGameRecord(winner: 'red' | 'black' | 'draw') {
  if (recordSaved) {
    console.log('记录已保存，跳过重复保存')
    return
  }
  recordSaved = true

  try {
    // 计算当前用户的结果：0=胜, 1=负, 2=和
    let result: number
    if (winner === 'draw') {
      result = 2
    } else if (winner === playerColor.value) {
      result = 0 // 胜
    } else {
      result = 1 // 负
    }

    // 仅使用有效历史，输出后端期望的紧凑格式（每步4位：fromX fromY toX toY）
    // 悔棋被从 validMoveHistory 移除后，这里不再包含悔棋前的步数
    const finalHistory = validMoveHistory.value
      .map(move => `${move.from.x}${move.from.y}${move.to.x}${move.to.y}`)
      .join('')

    await saveGameRecord({
      is_red: playerColor.value === 'red',
      result,
      history: finalHistory,
      start_time: gameStartTime.value.toISOString(),
      ai_level: aiLevel.value,
    })
    console.log('人机对战记录已保存', { validMoveHistory: validMoveHistory.value, finalHistory })
  } catch (error: any) {
    console.error('保存人机对战记录失败:', error)
    // 不阻塞用户体验，静默失败
  }
}

function refreshUserProfile() {
  try {
    getProfile().then((resp: any) => {
      const data = resp && typeof resp === 'object' && 'data' in resp ? resp.data : resp
      if (data) userStore.setUser(data)
    }).catch(() => {})
  } catch {}
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

    // 保存最新局面，确保刷新后仍是玩家回合
    saveAIGameStateToSession()
  } catch (error) {
    console.error('AI走棋错误:', error)
    showMsg('AI走棋出错')
    aiThinking.value = false
  }
}

onMounted(() => {
  console.log('AIView mounted, route.query:', route.query)
  console.log('playerColor:', playerColor.value, 'aiLevel:', aiLevel.value)
  // 进入人机对局时，主动刷新资料以更新右上角场次/胜率
  try {
    getProfile().then((resp: any) => {
      const d = resp && typeof resp === 'object' && 'data' in resp ? resp.data : resp
      if (d) userStore.setUser(d)
    }).catch(() => {})
  } catch {}

  // 【问题2修复】尝试从 sessionStorage 恢复之前的对局状态
  const savedAIGameState = sessionStorage.getItem('aiGameState')
  let isRestoringState = false
  let savedBoardState: any = null

  if (savedAIGameState) {
    try {
      const state = JSON.parse(savedAIGameState)
      // 检查恢复的状态是否有效（对局未结束）
      if (!state.gameOver) {
        console.log('恢复之前的人机对战状态:', state)
        // 恢复游戏参数
        playerColor.value = state.playerColor
        aiLevel.value = state.aiLevel
        gameStartTime.value = new Date(state.gameStartTime)
        moveHistory.value = state.moveHistory
        validMoveHistory.value = state.validMoveHistory || []
        currentTurn.value = state.currentTurn
        lastMove.value = state.lastMove
        savedBoardState = state.boardState || null
        isRestoringState = true
      } else {
        // 对局已结束，清除缓存
        sessionStorage.removeItem('aiGameState')
      }
    } catch (e) {
      console.warn('恢复人机对战状态失败:', e)
      sessionStorage.removeItem('aiGameState')
    }
  }

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

  // 【问题2修复】如果恢复了之前的状态，需要恢复棋盘局面
  if (isRestoringState) {
    const restoreMoves = (savedBoardState?.moveHistoryList && savedBoardState.moveHistoryList.length > 0)
      ? savedBoardState.moveHistoryList
      : validMoveHistory.value.map((move, index) => ({
          from: move.from,
          to: move.to,
          currentRole: (index % 2 === 0 ? 'self' : 'enemy') as const,
          pieceName: move.pieceName,
          pieceColor: move.pieceColor,
        }))

    if (restoreMoves.length > 0) {
      console.log('恢复棋盘局面，共有', restoreMoves.length, '步')
      chessBoard.restoreState({
        isNetPlay: false,
        selfColor: playerColor.value,
        moveHistory: restoreMoves,
        currentRole: savedBoardState?.currentRole || 'self',
        mode: 'ai',
      })
      console.log('棋盘状态已恢复，chessBoard.stepsNum =', chessBoard.stepsNum)
    }
  } else {
    // 新对局开始时，清空所有历史记录
    moveHistory.value = ''
    validMoveHistory.value = []
  }

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
      // 【复盘修复】保存游戏状态到 gameStore，供复盘功能使用
      saveGameStateForReplay()
      // 清理会话缓存
      clearAIGameStateFromSession()
      // 保存人机对战记录并刷新资料
      saveAIGameRecord(winner)
      refreshUserProfile()
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
        // 【复盘修复】保存游戏状态到 gameStore，供复盘功能使用
        saveGameStateForReplay()
        // 清理会话缓存
        clearAIGameStateFromSession()
        // 保存人机对战记录并刷新资料
        saveAIGameRecord(winner)
        refreshUserProfile()
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
      // 【复盘修复】保存游戏状态到 gameStore，供复盘功能使用
      saveGameStateForReplay()
      // 清理会话缓存
      clearAIGameStateFromSession()
      // 保存人机对战记录并刷新资料
      saveAIGameRecord(winner)
      refreshUserProfile()
    }
  })

  // 初始化当前回合与最近落子显示
  try {
    currentTurn.value = chessBoard.currentRole === 'self' ? '你的回合' : '对手回合'
    const mh = chessBoard.moveHistoryList || []
    moveCount.value = mh.length
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
    // 只在游戏进行中才记录步数，避免游戏结束后刷新导致重复计数
    if (gameOver.value) {
      console.log('游戏已结束，忽略 BOARD:MOVE:MADE 事件')
      return
    }

    lastMove.value = formatMoveLabel(from, to, pieceName, pieceColor)
    // 【问题1修复】维护两份历史：
    // 1. moveHistory：用于发送给后端的紧凑格式（需在悔棋时正确处理）
    // 2. validMoveHistory：用于计算最终的有效步数
    if (!isReplaying) {
      moveHistory.value += `${from.x}${from.y}${to.x}${to.y}`
      validMoveHistory.value.push({ from, to, pieceName, pieceColor })
      moveCount.value = validMoveHistory.value.length

      // 【问题2修复】落子后立即保存当前对局状态到 sessionStorage
      // 这样刷新页面时可以恢复对局
      saveAIGameStateToSession()
    }
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
      // 修复：触发 LOCAL:GAME:END 而不是 GAME:END，确保游戏正确结束并保存状态
      channel.emit('LOCAL:GAME:END', { winner: aiColor })
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
  // 仅在新对局且尚无步数时让AI先手；恢复对局时不触发，避免刷新多走
  if (!isRestoringState && playerColor.value === 'black' && chessBoard.stepsNum === 0) {
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
            class="group relative px-6 py-2.5 bg-red-50 text-red-700 rounded-xl font-bold shadow-sm hover:bg-red-100 hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2 border border-red-100"
            @click="giveUp"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 21v-8a2 2 0 012-2h14a2 2 0 012 2v8M3 21h18M5 21v-8a2 2 0 012-2h14a2 2 0 012 2v8" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            认输
          </button>
          <button
            class="group relative px-6 py-2.5 bg-gray-100 text-gray-700 rounded-xl font-bold shadow-sm hover:bg-gray-200 hover:shadow-md hover:-translate-y-0.5 transition-all duration-200 flex items-center gap-2 border border-gray-200"
            @click="quit"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            退出
          </button>
        </div>
      </div>

      <!-- 右侧：信息面板 -->
      <div class="w-full sm:w-72 lg:w-80 flex-none flex flex-col gap-3 h-auto sm:h-full overflow-y-auto">
        <!-- AI对战信息面板 -->
        <div class="bg-white/60 backdrop-blur-md rounded-2xl shadow-sm p-3 border border-white/50 flex flex-col">
          <div class="flex items-center justify-between w-full mb-3">
            <!-- 玩家 -->
            <div class="flex flex-col items-center w-1/3 group">
              <div class="relative">
                <img :src="userStore.userInfo?.avatar || '/images/default_avatar.png'" alt="玩家头像" class="w-12 h-12 rounded-full mb-1 object-cover border-4 shadow-md transition-transform group-hover:scale-105" :class="playerColor === 'red' ? 'border-red-500' : 'border-gray-800'" />
                <div class="absolute -bottom-1 -right-1 w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold text-white shadow-sm" :class="playerColor === 'red' ? 'bg-red-500' : 'bg-gray-800'">
                  {{ playerColor === 'red' ? '红' : '黑' }}
                </div>
              </div>
              <span class="text-xs truncate w-full text-center font-bold text-amber-900">{{ userStore.userInfo?.name }}</span>
            </div>

            <!-- VS -->
            <div class="flex flex-col items-center justify-center">
              <span class="text-3xl font-black text-amber-200/80 italic">VS</span>
            </div>

            <!-- 电脑 -->
            <div class="flex flex-col items-center w-1/3 group">
              <div class="relative">
                <div class="w-12 h-12 rounded-full mb-1 bg-gradient-to-br from-indigo-500 to-blue-600 flex items-center justify-center text-white font-bold text-xl shadow-md border-4 border-white transition-transform group-hover:scale-105">
                  🤖
                </div>
                <div class="absolute -bottom-1 -right-1 w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold text-white shadow-sm" :class="playerColor === 'red' ? 'bg-gray-800' : 'bg-red-500'">
                  {{ playerColor === 'red' ? '黑' : '红' }}
                </div>
              </div>
              <span class="text-xs truncate w-full text-center font-bold text-amber-900">电脑 ({{ aiLabel }})</span>
            </div>
          </div>

          <!-- 数据展示 -->
          <div class="flex justify-between w-full gap-2 text-[10px] sm:text-xs">
            <div class="flex-1 bg-amber-50/50 p-2 rounded-lg border border-amber-100 flex flex-col gap-1">
              <div class="flex justify-between">
                <span class="text-amber-800/60">胜率</span>
                <span class="font-bold text-amber-900">{{ (userStore.userInfo?.winRate || 0).toFixed(0) }}%</span>
              </div>
              <div class="flex justify-between">
                <span class="text-amber-800/60">场次</span>
                <span class="font-bold text-amber-900">{{ userStore.userInfo?.totalGames || 0 }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-amber-800/60">经验</span>
                <span class="font-bold text-amber-900">{{ userStore.userInfo?.exp || 0 }}</span>
              </div>
            </div>
            <div class="flex-1 bg-indigo-50/50 p-2 rounded-lg border border-indigo-100 flex flex-col gap-1">
              <div class="flex justify-between">
                <span class="text-indigo-800/60">难度</span>
                <span class="font-bold text-indigo-900">{{ aiLabel }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-indigo-800/60">类型</span>
                <span class="font-bold text-indigo-900">人机对战</span>
              </div>
              <div class="flex justify-between">
                <span class="text-indigo-800/60">棋手</span>
                <span class="font-bold text-indigo-900">超级大脑</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 游戏统计 -->
        <div class="bg-white/60 backdrop-blur-md border border-white/50 rounded-2xl p-4 shadow-sm">
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
              <span class="font-bold text-amber-900 text-lg">人机对战</span>
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
    :visible="endModalVisible"
    :result="endResult"
    :on-review="() => router.push('/game/replay')"
    :on-quit="quit"
    @close="endModalVisible = false"
  />
  <ConfirmModal
    :visible="resignConfirmVisible"
    title="确认认输？"
    message="认输后本局将判负，是否继续？"
    confirmText="认输"
    cancelText="取消"
    :on-confirm="handleResignConfirm"
    :on-cancel="handleResignCancel"
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
