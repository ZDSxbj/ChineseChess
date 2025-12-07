import instance from '../useRequest'

export function checkPassword(data: { password: string }) {
  return instance.post('/user/check_password', data)
}
