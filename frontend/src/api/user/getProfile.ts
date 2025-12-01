import instance from '../useRequest'
import type { UserInfo } from './user'

/**
 * 请求当前登录用户的个人信息
 * @returns Promise<UserInfo> 用户信息对象
 */
export function getProfile() {
  // 先检查instance是否存在，避免因实例未定义导致的错误
  if (!instance) {
    throw new Error('请求实例未初始化，请检查useRequest配置')
  }

  return instance.get<UserInfo>('/user/profile', {
    withCredentials: true,
    timeout: 10000
  }).catch(error => {
    // 确保error对象存在
    if (!error) {
      throw new Error('获取个人信息失败：未知错误')
    }

    let errorMsg = '获取个人信息失败'
     
    // 安全处理error.response
    if (error.response) {
      const status = error.response.status || '未知状态码'
      const responseData = error.response.data || {}
      const message = responseData.message || '无错误信息'
      errorMsg = `请求错误（${status}）：${message}`
    } 
    // 处理无响应的情况（网络错误）
    else if (error.request) {
      errorMsg = '网络连接失败，请检查网络设置'
    } 
    // 处理请求配置错误
    else {
      // 确保error.message存在
      errorMsg = `请求配置错误：${error.message || '未知配置问题'}`
    }
    
    throw new Error(errorMsg)
  })
}