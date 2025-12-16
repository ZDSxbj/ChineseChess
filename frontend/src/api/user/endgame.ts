import RequestHandler from '../useRequest'

export interface EndgameCompleteReq {
  scenarioId: string
  difficulty: '初级' | '中级' | '高级' | string
  success: boolean
}

export interface EndgameCompleteResp {
  awarded: number
  alreadyAwarded: boolean
}

export function reportEndgameComplete(data: EndgameCompleteReq) {
  return RequestHandler.post<EndgameCompleteResp>('/user/endgame/complete', data)
}
