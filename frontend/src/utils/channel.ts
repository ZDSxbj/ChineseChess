type color = 'red' | 'black'
interface position { x: number, y: number }
interface GameEvents {
  'MATCH:SUCCESS': null
  'GAME:START': {
    color: color
  }
  'GAME:END': {
    winner: color
  }
  'NET:GAME:START': {
    color: color
  }
  'NET:GAME:END': {
    winner: color
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
  }
}

type Listener<T> = (req: T) => void

class Channel {
  private eventsQueue: {
    [K in keyof GameEvents]: Array<GameEvents[K]>
  } = {} as {
    [K in keyof GameEvents]: Array<GameEvents[K]>
  }

  private listeners: {
    [K in keyof GameEvents]?: Listener<GameEvents[K]>
  } = {}

  on<K extends keyof GameEvents>(eventName: K, listener: Listener<GameEvents[K]>) {
    if (!this.eventsQueue[eventName]) {
      this.eventsQueue[eventName] = []
    }
    this.listeners[eventName] = listener as Listener<GameEvents[keyof GameEvents]>
    while (this.eventsQueue[eventName]?.length) {
      const req = this.eventsQueue[eventName].shift()
      if (req === undefined) {
        continue
      }
      listener(req)
    }
  }

  off(eventName: keyof GameEvents) {
    if (!this.listeners[eventName])
      return
    delete this.listeners[eventName]
  }

  emit<K extends keyof GameEvents>(eventName: K, req: GameEvents[K]) {
    if (!this.listeners[eventName]) {
      if (!this.eventsQueue[eventName]) {
        this.eventsQueue[eventName] = []
      }
      this.eventsQueue[eventName].push(req)
      return
    }
    this.listeners[eventName](req)
  }
}

export default new Channel()
