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
