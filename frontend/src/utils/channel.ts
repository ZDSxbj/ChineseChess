type color = 'red' | 'black'
interface position { x: number, y: number }
interface GameEvents {
  'MATCH:SUCCESS': null
  'GAME:START': {
    color: color
  }
  'GAME:END': {
    winner: color
    online?: boolean
  }
  'GAME:MOVE': Record<string, never> // AI对战中玩家移动完成事件
  'LOCAL:GAME:END': {
    winner: color
  }
  'NET:GAME:START': {
    color: color
  }
  'NET:GAME:END': {
    winner: string
  }
  'NET:GAME:SYNC': {
    history: any[]
    role?: string
    currentTurn?: string
  }
  'NET:DRAW:REQUEST': Record<string, never>
  'NET:DRAW:RESPONSE': {
    accepted: boolean
  }
  'NET:CHESS:MOVE': {
    from: position
    to: position
  }
  'NET:CHESS:MOVE:END': {
    from: position
    to: position
  }
  'NET:CHESS:REGRET:REQUEST': Record<string, never> // 事件参数为空对象
  'NET:CHESS:REGRET:RESPONSE': {
    accepted: boolean
  }
  'NET:CHESS:REGRET:SUCCESS': Record<string, never>
  'NET:CHAT:MESSAGE': {
    sender: string
    content: string
    relationId?: number
    senderId?: number
    messageId?: number
    createdAt?: number
  }
  'NET:FRIEND:REQUEST': {
    requestId?: number
    senderId?: number
    senderName?: string
    content?: string
    createdAt?: number
  }
}

type Listener<T> = (req: T) => void

class Channel {
  private eventsQueue: {
    [K in keyof GameEvents]?: Array<GameEvents[K]>
  } = {} as {
    [K in keyof GameEvents]?: Array<GameEvents[K]>
  }

  private listeners: {
    [K in keyof GameEvents]?: Set<Listener<GameEvents[K]>>
  } = {}

  on<K extends keyof GameEvents>(eventName: K, listener: Listener<GameEvents[K]>) {
    if (!this.eventsQueue[eventName]) {
      this.eventsQueue[eventName] = []
    }
    if (!this.listeners[eventName]) {
      this.listeners[eventName] = new Set<Listener<GameEvents[K]>>() as Set<Listener<GameEvents[keyof GameEvents]>>
    }
    this.listeners[eventName]!.add(listener as Listener<GameEvents[keyof GameEvents]>)
    // drain queued events and deliver to this new listener
    const queue = this.eventsQueue[eventName] || []
    while (queue.length) {
      const req = queue.shift()
      if (req === undefined) continue
      try { listener(req) } catch { /* ignore listener errors */ }
    }
  }

  off<K extends keyof GameEvents>(eventName: K, listener?: Listener<GameEvents[K]>) {
    const set = this.listeners[eventName]
    if (!set)
      return
    if (listener === undefined) {
      // remove all listeners for this event
      delete this.listeners[eventName]
      return
    }
    set.delete(listener as Listener<GameEvents[keyof GameEvents]>)
    if (set.size === 0) {
      delete this.listeners[eventName]
    }
  }

  emit<K extends keyof GameEvents>(eventName: K, req: GameEvents[K]) {
    const set = this.listeners[eventName]
    if (!set || set.size === 0) {
      if (!this.eventsQueue[eventName]) {
        this.eventsQueue[eventName] = []
      }
      this.eventsQueue[eventName]!.push(req)
      return
    }
    // 调试：当发生游戏结束相关事件时打印调用栈，便于定位误触发来源
    if (eventName === 'GAME:END' || eventName === 'LOCAL:GAME:END' || eventName === 'NET:GAME:END') {
      try {
        // 尽量少量打印以避免控制台噪音
        // @ts-ignore
        console.groupCollapsed && console.groupCollapsed(`[Channel] emit ${String(eventName)}`)
        console.log('[Channel] emit payload:', req)
        console.trace('[Channel] emit stack trace')
        console.groupEnd && console.groupEnd()
      } catch (e) {
        console.log('[Channel] emit debug failed', e)
      }
    }
    // call all listeners
    for (const l of Array.from(set)) {
      try { l(req) } catch { /* ignore listener errors */ }
    }
  }

  // 新增：清空指定事件的队列
  clearQueue(eventName: keyof GameEvents) {
    if (this.eventsQueue[eventName]) {
      this.eventsQueue[eventName] = []
    }
  }
}

export default new Channel()
