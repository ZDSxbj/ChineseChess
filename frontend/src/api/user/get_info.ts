import instance from '../useRequest'

export interface GetInfoRequest {
  id: number
}

export interface GetInfoResponse {
  id: number
  name: string
}

export function login(req: GetInfoRequest) {
  return instance.post<GetInfoResponse>('/public/login', req)
}
