import type { AxiosRequestConfig } from 'axios'
import axios from 'axios'

import { useUserStore } from '@/store/useStore'
import apiBus from '@/utils/apiBus'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

const instance = axios.create({
  baseURL: API_URL,
  timeout: 5000,
})
interface Response<T = any> {
  code: number
  message: string
  data: T
}

type ErrorHandler = (resp?: Response) => void

instance.interceptors.request.use((config) => {
  const { url } = config
  if (!url || url.startsWith('/public')) {
    // 登录接口
    return config
  }
  // 获取token操作
  const { token } = useUserStore()
  // token 是一个 computed ref，取 .value 才是实际字符串
  const tokenValue = (token && typeof (token as any).value !== 'undefined') ? (token as any).value : token
  if (tokenValue) {
    config.headers.Authorization = `Bearer ${tokenValue}`
  }
  return config
})

const errorCodeHandler: Record<number, ErrorHandler> = {
  // - token无效或没有token 1 (前端需要重新登录)
  0: (resp?: Response) => {
    const message = resp?.message || 'fail'
    apiBus.emit('API:FAIL', { message })
  },
  1: () => {
    apiBus.emit('API:UN_AUTH', null)
  },
}

const httpCodeHandler: Record<number, ErrorHandler> = {
  404: () => {
    apiBus.emit('API:NOT_FOUND', null)
  },
  500: () => {},
}

instance.interceptors.response.use(
  (resp) => {
    const { code, data, message } = resp.data as Response
    if (code < 100) {
      // 业务错误处理
      errorCodeHandler[code]?.(resp.data)
      return Promise.reject(message)
    }
    return data
  },
  (error) => {
    const { status } = error
    httpCodeHandler[status]?.()
    return Promise.reject(error)
  },
)

async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const resp = await instance.request<T>(config)
  return resp as T
}

const RequestHandler = {
  get: <T>(url: string, params?: Record<string, any>) => {
    return request<T>({ url, method: 'get', params })
  },
  post: <T>(url: string, data?: Record<string, any>) => {
    return request<T>({ url, method: 'post', data })
  },
  put: <T>(url: string, data?: Record<string, any>) => {
    return request<T>({ url, method: 'put', data })
  },
  delete: <T>(url: string, data?: Record<string, any>) => {
    return request<T>({ url, method: 'delete', data })
  },
}

export default RequestHandler
