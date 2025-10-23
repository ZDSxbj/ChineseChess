import type { UserInfo } from '../user/user'

export interface RoomInfo {
  id: number
  current: UserInfo
  next: UserInfo
}
