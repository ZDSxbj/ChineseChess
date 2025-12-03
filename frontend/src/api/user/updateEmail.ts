// frontend/src/api/user/updateEmail.ts
import instance from '../useRequest'

export function updateEmail(data: { email: string, code: string }) {
  return instance.post('/user/update_email', data)
}