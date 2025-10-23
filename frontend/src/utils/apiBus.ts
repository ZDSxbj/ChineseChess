interface ApiEvents {
  'API:UN_AUTH': null
  'API:NOT_FOUND': null
  'API:LOGOUT': null
  'API:FAIL': {
    message: string
  }
  'API:LOGIN': {
    token: string
    name: string
    avatar: string
    exp: number
    stop?: boolean
  }
}

type Listener<T> = (req: T) => void

class ApiBus {
  private listeners: {
    [K in keyof ApiEvents]?: Set<Listener<ApiEvents[K]>>
  } = {}

  on<K extends keyof ApiEvents>(eventName: K, listener: Listener<ApiEvents[K]>) {
    if (!this.listeners[eventName]) {
      this.listeners[eventName] = new Set<Listener<ApiEvents[keyof ApiEvents]>>()
    }
    this.listeners[eventName].add(listener)
  }

  off<K extends keyof ApiEvents>(eventName: K, listener: Listener<ApiEvents[K]>) {
    if (!this.listeners[eventName])
      return
    this.listeners[eventName].delete(listener)
  }

  emit<K extends keyof ApiEvents>(eventName: K, req: ApiEvents[K]) {
    this.listeners[eventName]?.forEach((listener) => {
      if (req === undefined)
        return
      listener(req)
    })
  }
}

export default new ApiBus()
