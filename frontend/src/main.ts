import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import './assets/main.css'
import 'virtual:uno.css'
import { API_URL } from '@/api/useRequest'
import { heartbeat } from '@/api/user/heartbeat'

const app = createApp(App)

app.use(router)
const pinia = createPinia()
app.use(pinia)

app.mount('#app')

// 在页面关闭/刷新时，尽力通知后端将在线状态置为离线（不阻塞页面关闭）
try {
	window.addEventListener('beforeunload', () => {
		// 读取 token（记住我优先 localStorage，其次 sessionStorage）
		const remember = typeof localStorage !== 'undefined' && localStorage.getItem('rememberMe') === '1'
		const storage = remember ? localStorage : sessionStorage
		const token = storage.getItem('token')
		if (!token) return
		const url = `${API_URL}/user/logout?token=${encodeURIComponent(token)}`
		// sendBeacon 使用 POST，无需自定义头，后端从 query 读取 token
		try {
			const blob = new Blob([], { type: 'text/plain' })
			navigator.sendBeacon(url, blob)
		} catch {
			// 兜底：使用 keepalive fetch（不会阻塞页面关闭）
			try {
				fetch(url, { method: 'POST', keepalive: true })
			} catch {}
		}
	})
} catch {}

// 周期性心跳：保持在线状态，并让后端有依据进行过期下线
try {
	// 延迟几秒启动，确保 Pinia/Store 已初始化
	setTimeout(() => {
		const interval = 30_000 // 30s
		// 立即心跳一次（已登录才会带上 Authorization）
		heartbeat().catch(() => {})
		setInterval(() => {
			heartbeat().catch(() => {})
		}, interval)
	}, 3000)
} catch {}
