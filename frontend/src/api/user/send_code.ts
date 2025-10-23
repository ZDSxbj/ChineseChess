import instance from '../useRequest'

export interface SendCodeRequest {
  email: string
}

export function sendCode(req: SendCodeRequest) {
  return instance.post('/public/send-code', req)
}
