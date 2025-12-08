import Request from '@/api/useRequest'

export interface FriendItem {
  id: number
  relationId?: number
  name: string
  avatar: string
  online: boolean
  gender: string
  exp: number
  totalGames: number
  winRate: number
  unreadCount?: number
  challengePending?: boolean
}

export async function getFriends(): Promise<{ friends: FriendItem[] }> {
  return Request.get('/user/friends')
}

export async function deleteFriend(friendId: number) {
  return Request.delete(`/user/friends/${friendId}`)
}

export async function getFriendRequests(): Promise<{ requests: any[] }> {
  return Request.get('/user/friend-requests')
}

export async function getFriendChallenges(): Promise<{ challenges: Array<{ id: number, friendId: number, senderId: number, receiverId: number, roomId?: number, createdAt: string }> }> {
  return Request.get('/user/friend-challenges')
}

export async function acceptFriendRequest(requestId: number) {
  return Request.post(`/user/friend-requests/${requestId}/accept`)
}

export async function deleteFriendRequest(requestId: number) {
  return Request.delete(`/user/friend-requests/${requestId}`)
}

export async function checkFriendRequest(receiverId: number): Promise<{ exists: boolean }> {
  return Request.get(`/user/friend-requests/check?receiverId=${receiverId}`)
}
