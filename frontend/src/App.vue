<script setup lang="ts">
import { onMounted, onUnmounted, provide, ref } from 'vue'
import { login } from './api/user/login'
import { showMsg } from './components/MessageBox'
import { useUserStore } from './store/useStore'
import apiBus from './utils/apiBus'
import { useWebSocket } from './websocket'

useUserStore()
apiBus.on('API:FAIL', (req) => {
  const { message } = req
  showMsg(message)
})

const ws = useWebSocket()
provide('ws', ws)

const WebsocketURL = import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8080/ws'

apiBus.on('API:LOGIN', (req) => {
  const { token } = req
  ws.connect(`${WebsocketURL}?token=${token}`)
})

const isPC = ref(false)
provide('isPC', isPC)

function handleResize() {
  isPC.value = window.innerWidth > 640
}

// 根据 rememberMe，从对应存储读取自动登录凭据
const rememberFlag = localStorage.getItem('rememberMe')
const storage = (rememberFlag === '1') ? localStorage : sessionStorage
const email = storage.getItem('email')
const password = storage.getItem('password')
if (email && password) {
  login({ email, password }).then((resp) => {
    apiBus.emit('API:LOGIN', { ...resp, stop: true })
  })
}
onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
})
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <RouterView />
</template>
