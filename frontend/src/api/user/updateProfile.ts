import type { UserInfo } from './user'
import RequestHandler from '../useRequest'

export function updateProfile(data: Partial<UserInfo>) {
  return RequestHandler.post('/user/profile', data, {
    withCredentials: true,
    timeout: 10000,
  })
}
