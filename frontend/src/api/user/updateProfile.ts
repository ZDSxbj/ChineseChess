import RequestHandler from '../useRequest'
import type { UserInfo } from './user'

export function updateProfile(data: Partial<UserInfo>) {
  return RequestHandler.post('/user/profile', data, {
    withCredentials: true,
    timeout: 10000
  })
}