import RequestHandler from '../useRequest'

export function heartbeat() {
  return RequestHandler.post('/user/heartbeat')
}
