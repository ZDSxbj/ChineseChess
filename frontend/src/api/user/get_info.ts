import instance from '../useRequest'
import type { UserInfo } from './user'

export interface GetInfoRequest {
  id?: number
  name?: string
}

// The backend returns the UserInfo fields directly in the data object because of struct embedding
export type GetInfoResponse = UserInfo

export function getUserInfo(req: GetInfoRequest) {
  return instance.post<GetInfoResponse>('/info', req)
}
