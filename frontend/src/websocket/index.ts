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
  CancelMatch: 9, // 取消匹配
  Error: 10,
  RegretRequest: 11, // 悔棋请求
  RegretResponse: 12, // 悔棋响应
  DrawRequest: 13, // 和棋请求
  DrawResponse: 14, // 和棋响应
  ChatMessage: 15, // 聊天消息
  Sync: 16, // 同步房间状态（重连时服务端发送）
  FriendRequest: 17, // 好友申请
  FriendChallengeInvite: 18, // 好友对弈邀请
  FriendChallengeCancel: 19, // 撤销对弈邀请
  FriendChallengeAccept: 20, // 接受对弈邀请
  FriendChallengeReject: 21, // 拒绝对弈邀请
  FriendChallengeCreated: 22, // 邀请已创建（仅发给发送者，含 challengeId、roomId）
  AIMove: 23, // AI走棋
  CreateAI: 24, // 创建AI房间
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
  relationId?: number
  senderId?: number
  messageId?: number
  createdAt?: number
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
  cancelMatch: () => void // 【问题2修复】新增取消匹配方法
  join: (id: number) => void
  create: () => Promise<unknown>
  giveUp: () => void
  sendRegretRequest: () => void
  sendRegretResponse: (accepted: boolean) => void
  sendChatMessage: (content: string) => void // 新增发送聊天消息方法
  sendFriendRequest: (receiverId: number, content: string) => void
  sendDrawRequest: () => void
  sendDrawResponse: (accepted: boolean) => void
  getCurrentRoomId: () => number | undefined // 新增获取当前房间ID方法
  // 好友挑战相关
  sendFriendChallengeInvite: (receiverId: number, relationId: number) => void
  sendFriendChallengeCancel: (challengeId: number, receiverId: number) => void
  sendFriendChallengeAccept: (challengeId: number, senderId: number, roomId: number) => void
  sendFriendChallengeReject: (challengeId: number, senderId: number) => void
}

export function useWebSocket(): WebSocketService {
  const socket = ref<WebSocket | null>(null)
  let lastUrl: string | null = null
  const pendingMessages: WebSocketMessage[] = []
  const currentRoomId = ref<number | undefined>(undefined) // 新增：记录当前房间ID
  let resolve: (value?: unknown) => void
  // let reject: (reason?: unknown) => void

  const eventHandler: EventHandler = (data) => {
    const { type } = data
    switch (type) {
      case MessageType.Normal:
      { const { message } = data
        // 如果是房间不存在的错误，在游戏结束流程中会误触发该信息，过滤掉以免和胜利信息重复
        if (message === '房间不存在') {
          console.warn('Ignored message: 房间不存在')
          break
        }
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
      { 
        // 检查是否是AI游戏开始
        const payload = data as any
        if (payload.aiColor) {
          // AI游戏开始
          showMsg('AI对战开始')
          channel.emit('AI:GAME:START', { 
            color: payload.role, 
            opponent: payload.opponent,
            aiColor: payload.aiColor,
          })
        } else {
          // 普通游戏开始
          showMsg('Game started')
          const { role, opponent } = data as { role: 'red' | 'black', opponent: any }
          channel.emit('MATCH:SUCCESS', null)
          channel.emit('NET:GAME:START', { color: role, opponent })
        }
        break }
      case MessageType.End:
      { const { winner, reason } = data as any
        if (winner === undefined) {
          return
        }
        // 把后端的数值映射为应用内事件（'red'|'black'|'draw'）并通知 UI 层
        let w: string | undefined
        if (winner === 0)
          w = 'draw'
        else if (winner === 1)
          w = 'red'
        else if (winner === 2)
          w = 'black'
        if (w) {
          channel.emit('NET:GAME:END', { winner: w, reason: reason || '' })
        }
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
      case MessageType.DrawRequest:
        // 收到和棋请求，通知 UI
        channel.emit('NET:DRAW:REQUEST', {})
        break
      case MessageType.RegretResponse: {
        // 处理悔棋响应
        const accepted = data.accepted
        channel.emit('NET:CHESS:REGRET:RESPONSE', { accepted })
        break
      }
      case MessageType.FriendChallengeCreated: {
        const payload = data as any
        channel.emit('NET:FRIEND:CHALLENGE:CREATED', payload)
        break
      }
      case MessageType.DrawResponse: {
        // 处理和棋响应
        const accepted = data.accepted
        channel.emit('NET:DRAW:RESPONSE', { accepted })
        break
      }
      case MessageType.ChatMessage: {
        // 处理聊天消息，尽量使用 relationId/senderId 做精确路由
        const { sender, content, relationId, senderId, messageId, createdAt } = data as any
        if (content) {
          channel.emit('NET:CHAT:MESSAGE', { sender, content, relationId, senderId, messageId, createdAt })
        }
        break
      }
      case MessageType.FriendRequest: {
        const payload = data as any
        // 转发到 channel，社交界面可订阅处理
        channel.emit('NET:FRIEND:REQUEST', {
          requestId: payload.requestId,
          senderId: payload.senderId,
          senderName: payload.senderName,
          content: payload.content,
          createdAt: payload.createdAt,
        })
        break
      }
      case MessageType.FriendChallengeInvite: {
        const payload = data as any
        // 全局提示
        showMsg(`收到了 ${payload.senderName || '好友'} 的对弈邀请，请及时在社交界面中查收`)
        channel.emit('NET:FRIEND:CHALLENGE:INVITE', payload)
        break
      }
      case MessageType.FriendChallengeCancel: {
        const payload = data as any
        showMsg('对方撤销了好友对弈邀请')
        channel.emit('NET:FRIEND:CHALLENGE:CANCEL', payload)
        break
      }
      case MessageType.FriendChallengeAccept: {
        const payload = data as any
        channel.emit('NET:FRIEND:CHALLENGE:ACCEPT', payload)
        break
      }
      case MessageType.FriendChallengeReject: {
        const payload = data as any
        showMsg('对方拒绝了你的对弈邀请')
        channel.emit('NET:FRIEND:CHALLENGE:REJECT', payload)
        break
      }
      case MessageType.Sync: {
        // 服务端发送的重连同步消息，转发到事件通道
        // 格式参考后端: { history: Position[], role: 'red'|'black', currentTurn: 'red'|'black' }
        const payload = data as any
        channel.emit('NET:GAME:SYNC', { history: payload.history || [], role: payload.role, currentTurn: payload.currentTurn })
        break
      }
      case MessageType.AIMove: {
        // AI走棋消息
        let { from, to } = data
        from = translateChessPosition(from!)
        to = translateChessPosition(to!)
        if (from && to) {
          channel.emit('AI:MOVE', { from, to })
        }
        break
      }
      default:
        break
    }
  }

  const sendMessage = (message: WebSocketMessage) => {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      console.log('[WebSocket 发送]', message)
      socket.value.send(JSON.stringify(message))
      return
    }
    // 如果 socket 不可用，尝试使用上次的 URL 重连并排队消息
    console.warn('WebSocket not open - queuing message and attempting reconnect', socket.value?.readyState)
    showMsg('WebSocket 未连接，正在尝试重连...')
    if (lastUrl) {
      pendingMessages.push(message)
      try {
        // eslint-disable-next-line ts/no-use-before-define
        connect(lastUrl)
      }
      catch (e) {
        console.error('Reconnect failed', e)
        showMsg('重连失败，请检查网络或刷新页面')
      }
    }
    else {
      showMsg('WebSocket 未连接，无法发送消息')
    }
  }

  const match = () => {
    sendMessage({
      type: MessageType.Match,
    })
  }

  // 【问题2修复】实现取消匹配方法
  const cancelMatch = () => {
    sendMessage({
      type: MessageType.CancelMatch,
    })
  }

  const move = (from: ChessPosition, to: ChessPosition) => {
    sendMessage({
      type: MessageType.Move,
      from,
      to,
    })
  }

  const end = (winner?: string) => {
    const payload: any = { type: MessageType.End }
    if (winner) {
      // 将字符串颜色映射为后端约定的数字：1 红胜, 2 黑胜
      const mapping: Record<string, number> = { red: 1, black: 2 }
      const w = mapping[winner]
      if (w)
        payload.winner = w
    }
    sendMessage(payload)
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
      reason: 'resign',
    } as any)
  }

  // 发送和棋请求
  const sendDrawRequest = () => {
    sendMessage({
      type: MessageType.DrawRequest,
    })
  }

  // 发送和棋响应
  const sendDrawResponse = (accepted: boolean) => {
    sendMessage({
      type: MessageType.DrawResponse,
      accepted,
    })
  }

  const connect = (url: string) => {
    lastUrl = url
    socket.value = new WebSocket(url)

    socket.value.onopen = () => {
      console.log('WebSocket connection opened.')
      channel.on('NET:CHESS:MOVE:END', ({ from, to }) => {
        move(from, to)
      })
      channel.on('GAME:END', (data: any) => {
        // data may contain { winner: 'red' | 'black', online: boolean }
        const isOnline = data && typeof data.online === 'boolean' ? data.online : true
        if (isOnline) {
          if (data && (data as any).winner) {
            end((data as any).winner)
          }
          else {
            // fallback: send end without winner
            end()
          }
        }
      })
      // flush pending messages
      while (pendingMessages.length > 0 && socket.value && socket.value.readyState === WebSocket.OPEN) {
        const msg = pendingMessages.shift()!
        console.log('[WebSocket 发送-队列]', msg)
        socket.value.send(JSON.stringify(msg))
      }
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

  // 发送好友申请
  const sendFriendRequest = (receiverId: number, content: string) => {
    sendMessage({
      type: MessageType.FriendRequest,
      receiverId,
      content,
    } as any)
  }

  // 发送好友对战挑战邀请
  const sendFriendChallengeInvite = (receiverId: number, relationId: number) => {
    sendMessage({
      type: MessageType.FriendChallengeInvite,
      receiverId,
      relationId,
    } as any)
  }

  // 撤销挑战
  const sendFriendChallengeCancel = (challengeId: number, receiverId: number) => {
    sendMessage({
      type: MessageType.FriendChallengeCancel,
      challengeId,
      receiverId,
    } as any)
  }

  // 接受挑战
  const sendFriendChallengeAccept = (challengeId: number, senderId: number, roomId: number) => {
    sendMessage({
      type: MessageType.FriendChallengeAccept,
      challengeId,
      senderId,
      roomId,
    } as any)
  }

  // 拒绝挑战
  const sendFriendChallengeReject = (challengeId: number, senderId: number) => {
    sendMessage({
      type: MessageType.FriendChallengeReject,
      challengeId,
      senderId,
    } as any)
  }

  // 新增：获取当前房间ID的方法
  const getCurrentRoomId = () => currentRoomId.value

  return {
    connect,
    close,
    end,
    match,
    cancelMatch, // 【问题2修复】导出取消匹配方法
    move,
    join,
    create,
    giveUp,
    sendRegretRequest,
    sendRegretResponse,
    sendChatMessage,
    sendDrawRequest,
    sendDrawResponse,
    getCurrentRoomId, // 新增导出获取当前房间ID的方法
    sendFriendRequest,
    sendFriendChallengeInvite,
    sendFriendChallengeCancel,
    sendFriendChallengeAccept,
    sendFriendChallengeReject,
  }
}
