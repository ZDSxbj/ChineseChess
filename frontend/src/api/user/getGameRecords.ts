import instance from '../useRequest'

export interface GameRecordItem {
  id: number
  opponent_id: number
  opponent_name: string
  // result: 0=胜, 1=负, 2=和
  result: number
  // game_type: 0=随机匹配, 1=人机对战, 2=好友对战
  game_type: number
  // is_red: true=红方, false=黑方
  is_red: boolean
  total_steps: number
  history: string
  start_time: string
}

export interface GetGameRecordsResponse {
  records: GameRecordItem[]
}

/**
 * 获取当前登录用户的对局记录
 * @returns Promise<GetGameRecordsResponse> 对局记录列表
 */
export function getGameRecords() {
  return instance.get<GetGameRecordsResponse>('/user/game-records')
}
