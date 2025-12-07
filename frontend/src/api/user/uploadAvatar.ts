import RequestHandler from '../useRequest'

export interface UploadAvatarResp {
  path: string // 例如 "/uploads/avatars/u1_1700000000.png"
}

export function uploadAvatar(file: File) {
  const fd = new FormData()
  fd.append('avatar', file)
  // 后端返回 { code,message,data:{path} }，拦截器会返回 data 字段
  return RequestHandler.post<UploadAvatarResp>('/user/avatar', fd)
}
