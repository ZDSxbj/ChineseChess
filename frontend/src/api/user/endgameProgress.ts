import RequestHandler from '../useRequest'

export interface EndgameProgress {
  attempts: number
  bestSteps: number | null
}

export function getEndgameProgress(scenarioId: string) {
  return RequestHandler.get<EndgameProgress>('/user/endgame/progress', { scenarioId })
}

export function recordEndgameProgress(scenarioId: string, result: 'win'|'lose', steps?: number) {
  return RequestHandler.post<EndgameProgress>('/user/endgame/progress', { scenarioId, result, steps })
}
