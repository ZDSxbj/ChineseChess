import instance from '../useRequest'

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  name: string
  avatar: string
  exp: number
}

export function login(req: LoginRequest) {
  return instance.post<LoginResponse>('/public/login', req)
}
