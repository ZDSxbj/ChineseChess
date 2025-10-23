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
  'NET:CHESS:MOVE': {
    from: position
    to: position
  }
  'NET:CHESS:MOVE:END': {
    from: position
    to: position
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
