import instance from '../useRequest'

export function updatePassword(data: { oldPassword: string; newPassword: string }) {
  return instance.post('/user/update_password', data)
}
