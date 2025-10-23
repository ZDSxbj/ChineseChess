import type { RoomInfo } from './room'

import instance from '../useRequest'

export interface GetSpareRoomResponse {
  rooms: RoomInfo[]
}

export function getSpareRoom() {
  return instance.post<GetSpareRoomResponse>('/user/rooms')
}
