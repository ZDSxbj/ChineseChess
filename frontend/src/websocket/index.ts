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
      case MessageType.End:
      { const { winner } = data
        if (winner === undefined) {
          return
        }
        const result = {
          0: '和棋',
          1: '红胜',
          2: '黑胜',
        } as const
        showMsg(result[winner])
        break }
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
        // 收到悔悔棋请求给UI
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
  }
}
