import Request from '@/api/useRequest'

export interface ScenarioProgress {
  scenario_id: string
  attempts: number
  best_steps?: number
  last_result?: 'win' | 'lose'
}

export interface GetEndgameProgressResponse {
  progress: ScenarioProgress[]
}

// 获取当前登录用户的所有残局进度
export function getEndgameProgress() {
  return Request.get<GetEndgameProgressResponse>('/user/endgame/progress')
}

// 保存单个残局关卡的一次挑战结果
export function saveEndgameProgress(params: { scenarioId: string; result: 'win' | 'lose'; steps?: number }) {
  const payload: any = {
    scenario_id: params.scenarioId,
    result: params.result,
  }
  if (params.result === 'win' && typeof params.steps === 'number') {
    payload.steps = params.steps
  }
  return Request.post('/user/endgame/progress', payload)
}


