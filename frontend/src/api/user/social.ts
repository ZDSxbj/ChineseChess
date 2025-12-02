import Request from '@/api/useRequest'

export interface FriendItem {
  id: number
  name: string
  avatar: string
  online: boolean
  gender: string
  exp: number
  totalGames: number
  winRate: number
}

export const getFriends = async (): Promise<{ friends: FriendItem[] }> => {
  return Request.get('/user/friends')
}

export const deleteFriend = async (friendId: number) => {
  return Request.delete(`/user/friends/${friendId}`)
}
