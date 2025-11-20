import type { ChessPosition } from '@/composables/ChessPiece'
import { ref } from 'vue'
import { showMsg } from '@/components/MessageBox'
import channel from '@/utils/channel'

const MessageType = {
  Normal: 1,
  Match: 2,
  Move: 3,
  Start: 4,
  End: 5,
  Join: 6,
  Create: 7,
  GiveUp: 8,
  Error: 10,
  RegretRequest: 11, // 悔棋请求
  RegretResponse: 12, // 悔棋响应
  ChatMessage: 15, // 聊天消息
  DrawRequest: 13, // 新增：和棋请求
  DrawResponse: 14, // 新增：和棋响应
} as const

interface WebSocketMessage {
  type: (typeof MessageType)[keyof typeof MessageType]
  message?: string
  from?: ChessPosition
  to?: ChessPosition
  role?: string
  timestamp?: number
  winner?: 0 | 1 | 2
  roomId?: number
  accepted?: boolean // 用于悔棋响应的布尔值字段
  content?: string // 聊天消息内容
  sender?: string // 聊天消息发送者
}

function translateChessPosition(position: ChessPosition): ChessPosition {
  const { x, y } = position
  return {
    x: 8 - x,
    y: 9 - y,
  }
}

export type EventHandler = (data: WebSocketMessage) => void

export interface WebSocketService {
  connect: (url: string) => void
  close: () => void

  end: (winner: string) => void
  move: (from: ChessPosition, to: ChessPosition) => void
  match: () => void
  join: (id: number) => void
  create: () => Promise<unknown>
  giveUp: () => void
  sendRegretRequest: () => void
  sendRegretResponse: (accepted: boolean) => void
  sendChatMessage: (content: string) => void // 新增发送聊天消息方法
  getCurrentRoomId: () => number | undefined // 新增获取当前房间ID方法
  sendDrawRequest: () => void
  sendDrawResponse: (accepted: boolean) => void
}

export function useWebSocket(): WebSocketService {
  const socket = ref<WebSocket | null>(null)
  const currentRoomId = ref<number | undefined>(undefined) // 新增：记录当前房间ID
  let resolve: (value?: unknown) => void
  // let reject: (reason?: unknown) => void

  const eventHandler: EventHandler = (data) => {
    const { type } = data
    switch (type) {
      case MessageType.Normal:
      { const { message } = data
        showMsg(message?.toString() || '')
        break }
      case MessageType.Match:
        showMsg('Match started')
        break
      case MessageType.Move:
      { let { from, to } = data
        from = translateChessPosition(from!)
        to = translateChessPosition(to!)
        if (from && to) {
          channel.emit('NET:CHESS:MOVE', { from, to })
        }
        else {
          console.error('Invalid move data:', data)
        }
        break }
      case MessageType.Start:
      { showMsg('Game started')
        const { role } = data as { role: 'red' | 'black' }
        channel.emit('MATCH:SUCCESS', null)
        channel.emit('NET:GAME:START', { color: role })
        break }
      case MessageType.End: {
        const { winner } = data
        if (winner === undefined) {
          return
        }

        // 1. 映射winner到颜色和结果文本
        const winnerMap = {
          0: { color: null, text: '和棋' }, // 和棋
          1: { color: 'red', text: '红胜' }, // 红方胜利
          2: { color: 'black', text: '黑胜' }, // 黑方胜利
        } as const

        const { color: winnerColor, text: resultText } = winnerMap[winner]

        // 2. 判断是否为认输（关键逻辑）
        // 规则：如果是和棋（0），则不是认输；如果是1/2，需要根据业务推断是否为认输
        // （假设：认输导致的胜利会通过该消息返回，这里统一标记为isResign，实际可根据后端是否传参调整）
        const isResign = winner !== 0 // 非和棋的结束都可能是认输或将死，先统一标记（后续可细化）

        // 3. 发送事件给ChessView，附带详细信息
        channel.emit('GAME:END', {
          winner: winnerColor, // 胜利方颜色（red/black/null）
          isResign, // 是否为认输导致的结束
        })

        // 4. 显示基础消息（可根据是否为认输细化）
        if (isResign) {
          // 认输场景：明确提示“对方认输”
          const loserText = winner === 1 ? '黑方' : '红方'
          showMsg(`${loserText}已认输，${resultText}`)
        }
        else {
          // 和棋或正常胜利
          showMsg(resultText)
        }
        break
      }
      case MessageType.Error:
      { const { message: errorMessage } = data
        showMsg(errorMessage?.toString() || 'Error occurred')
        break }
      case MessageType.Join:
        showMsg('Joined room')
        break
      case MessageType.Create:
        resolve?.()
        break
      case MessageType.RegretRequest:
        // 收到悔棋请求给UI
        channel.emit('NET:CHESS:REGRET:REQUEST', {})
        break
      case MessageType.RegretResponse: {
        // 处理悔棋响应
        const accepted = data.accepted
        channel.emit('NET:CHESS:REGRET:RESPONSE', { accepted })
        break
      }
      case MessageType.ChatMessage: {
        // 处理聊天消息
        const { sender, content } = data
        if (sender && content) {
          channel.emit('NET:CHAT:MESSAGE', { sender, content })
        }
        break
      }
      case MessageType.DrawRequest:
        // 新增：收到和棋请求给UI
        channel.emit('NET:CHESS:DRAW:REQUEST', {})
        break
      case MessageType.DrawResponse: {
        // 新增：处理和棋响应
        const accepted = data.accepted
        channel.emit('NET:CHESS:DRAW:RESPONSE', { accepted })
        break
      }
      default:
        break
    }
  }

  const sendMessage = (message: WebSocketMessage) => {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      // 打印发送的消息
      console.log('[WebSocket 发送]', message)
      socket.value.send(JSON.stringify(message))
    }
    else {
      console.log(socket.value, socket.value?.readyState)
      showMsg('WebSocket is not open. Unable to send message.')
    }
  }

  const match = () => {
    sendMessage({
      type: MessageType.Match,
    })
  }

  const move = (from: ChessPosition, to: ChessPosition) => {
    sendMessage({
      type: MessageType.Move,
      from,
      to,
    })
  }

  const end = () => {
    sendMessage({
      type: MessageType.End,
    })
  }

  const join = (roomId: number) => {
    currentRoomId.value = roomId // 记录当前房间ID
    sendMessage({
      type: MessageType.Join,
      roomId,
    })
  }

  const create = () => {
    sendMessage({
      type: MessageType.Create,
    })
    return new Promise((resolvePromise) => {
      resolve = resolvePromise
    })
  }

  const giveUp = () => {
    sendMessage({
      type: MessageType.GiveUp,
    })
  }

  const connect = (url: string) => {
    socket.value = new WebSocket(url)

    socket.value.onopen = () => {
      console.log('WebSocket connection opened.')
      channel.on('NET:CHESS:MOVE:END', ({ from, to }) => {
        move(from, to)
      })
      channel.on('GAME:END', () => {
        end()
      })
    }

    socket.value.onmessage = (event) => {
      // 打印原始JSON字符串
      console.log('[WebSocket 接收] 原始消息:', event.data)
      // 解析后的数据
      const data = JSON.parse(event.data)
      console.log('[WebSocket 接收] 解析后消息:', data)
      eventHandler(JSON.parse(event.data))
    }

    socket.value.onclose = () => {
      console.log('WebSocket connection closed.')
    }

    socket.value.onerror = (error) => {
      console.log('WebSocket error:', error)
    }
  }

  const close = () => {
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }
  }

  // 发送悔棋请求
  const sendRegretRequest = () => {
    sendMessage({
      type: MessageType.RegretRequest,
    })
  }

  // 发送悔棋响应
  const sendRegretResponse = (accepted: boolean) => {
    sendMessage({
      type: MessageType.RegretResponse,
      accepted,
    })
  }

  // 发送聊天消息
  const sendChatMessage = (content: string) => {
    sendMessage({
      type: MessageType.ChatMessage,
      content,
    })
  }

  // 新增：获取当前房间ID的方法
  const getCurrentRoomId = () => currentRoomId.value

  // 新增：发送和棋响应
  const sendDrawResponse = (accepted: boolean) => {
    sendMessage({
      type: MessageType.DrawResponse,
      accepted,
    })
  }
  // 新增：发送和棋请求
  const sendDrawRequest = () => {
    sendMessage({
      type: MessageType.DrawRequest,
    })
  }
  return {
    connect,
    close,
    end,
    match,
    move,
    join,
    create,
    giveUp,
    sendRegretRequest,
    sendRegretResponse,
    sendChatMessage,
    getCurrentRoomId, // 新增导出获取当前房间ID的方法
    sendDrawRequest,
    sendDrawResponse,
  }
}
