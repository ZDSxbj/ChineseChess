import Request from '@/api/useRequest'

export interface EndgameCompleteRequest {
  scenarioId: string
  difficulty: '初级' | '中级' | '高级'
  success: boolean
}

export interface EndgameCompleteResponse {
  awarded: number
  alreadyAwarded: boolean
}

// 残局挑战结算：仅在首次通关按难度奖励经验（中级50/高级100）
export function reportEndgameComplete(payload: EndgameCompleteRequest) {
  return Request.post<EndgameCompleteResponse>('/user/endgame/complete', payload)
}
