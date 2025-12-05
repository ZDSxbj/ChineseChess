import Request from '@/api/useRequest'

export interface ChatMessage {
  id: number
  friendId: number
  senderId: number
  receiverId: number
  isRead: boolean
  content: string
  createdAt: string
}

export async function getMessages(relationId: number, limit = 100, offset = 0): Promise<{ messages: ChatMessage[] } | ChatMessage[]> {
  // 返回的后端为数组或统一包裹，根据后端实现而异
  return Request.get(`/user/friends/${relationId}/messages`)
}

export async function sendMessage(relationId: number, content: string) {
  return Request.post(`/user/friends/${relationId}/messages`, { content })
}

export async function markRead(relationId: number) {
  return Request.post(`/user/friends/${relationId}/mark-read`, {})
}
