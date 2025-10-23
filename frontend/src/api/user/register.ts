import instance from '../useRequest'

export interface RegisterRequest {
  name: string
  email: string
  password: string
  password2: string
  vcode: string
}

export interface RegisterResponse {
  token: string
  name: string
  avatar: string
  exp: number
}

export function register(req: RegisterRequest) {
  return instance.post<RegisterResponse>('/public/register', req)
}
