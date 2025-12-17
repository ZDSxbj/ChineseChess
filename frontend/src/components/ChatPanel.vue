<script lang="ts" setup>
import type { WebSocketService } from '@/websocket'
import { nextTick, ref } from 'vue'

const props = defineProps<{
  ws: WebSocketService
}>()

const messages = ref<Array<{ sender: string, content: string, time: string }>>([])
const inputMessage = ref('')
const maxLen = 30
const isMinimized = ref(false)
const chatScrollRef = ref<HTMLDivElement>()

async function scrollToBottom() {
  await nextTick()
  if (chatScrollRef.value) {
    chatScrollRef.value.scrollTop = chatScrollRef.value.scrollHeight
  }
}

function sendMessage() {
  const text = inputMessage.value.trim()
  if (!text)
    return
  props.ws.sendChatMessage(text)
  messages.value.push({ sender: '我', content: text, time: new Date().toLocaleTimeString() })
  inputMessage.value = ''
  scrollToBottom()
}

// 暴露接收消息的方法给父组件
defineExpose({
  receiveMessage: (sender: string, content: string) => {
    messages.value.push({ sender, content, time: new Date().toLocaleTimeString() })
    scrollToBottom()
  },
})
</script>

<template>
  <div class="w-full flex flex-col bg-[#fdf6e3] rounded-xl shadow-lg border border-amber-200 overflow-hidden transition-all duration-300" :class="[isMinimized ? 'h-12' : 'h-chat']">
    <div class="px-4 py-3 border-b border-amber-200 bg-amber-100/50 text-amber-900 font-bold flex justify-between items-center cursor-pointer" @click="isMinimized = !isMinimized">
      <span class="flex items-center gap-2 select-none">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
        聊天
      </span>
      <button
        class="w-6 h-6 flex items-center justify-center text-amber-700 hover:text-amber-900 focus:outline-none transition-transform hover:scale-110"
      >
        <svg v-if="isMinimized" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
        </svg>
      </button>
    </div>

    <div v-show="!isMinimized" ref="chatScrollRef" class="flex-1 p-4 overflow-y-auto chat-scroll bg-amber-50/30">
      <div v-for="(msg, index) in messages" :key="index" class="mb-4">
        <div class="message-row" :class="[msg.sender === '我' ? 'self' : 'other']">
          <div class="bubble shadow-sm" :class="[msg.sender === '我' ? 'bubble--self bg-amber-600 text-white' : 'bubble--other bg-white text-gray-800 border border-amber-100']">
            <div class="bubble-content text-sm">{{ msg.content }}</div>
            <div class="bubble-time text-[10px] opacity-70 mt-1 text-right" :class="{'text-amber-100': msg.sender === '我', 'text-gray-400': msg.sender !== '我'}">{{ msg.time }}</div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="!isMinimized" class="px-3 py-3 border-t border-amber-200 bg-white/50 chat-input">
      <div class="flex items-center space-x-2">
        <input
          v-model="inputMessage"
          type="text"
          :maxlength="maxLen"
          class="flex-1 px-3 py-2 border border-amber-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-amber-500 bg-white/80"
          placeholder="请输入信息..."
          @keyup.enter="sendMessage"
        />
        <div class="text-xs text-amber-700/50 w-8 text-center">{{ inputMessage.length }}/{{ maxLen }}</div>
        <button
          class="px-4 py-2 bg-amber-700 text-white rounded-lg hover:bg-amber-800 text-sm font-bold shadow-md transition-colors"
          @click="sendMessage"
        >
          发送
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.h-chat {
  height: 100%;
}
.chat-scroll {
  flex: 1 1 0%;
  min-height: 0;
  overflow-y: auto;
}

.chat-input {
  flex: 0 0 auto;
}

.message-row {
  display: flex;
}
.message-row.self {
  justify-content: flex-end;
}
.message-row.other {
  justify-content: flex-start;
}

.bubble {
  max-width: 75%;
  padding: 8px 12px;
  border-radius: 12px;
}
.bubble--self {
  border-bottom-right-radius: 2px;
}
.bubble--other {
  border-bottom-left-radius: 2px;
}

.bubble-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.chat-scroll::-webkit-scrollbar {
  width: 6px;
}
.chat-scroll::-webkit-scrollbar-track {
  background: transparent;
}
.chat-scroll::-webkit-scrollbar-thumb {
  background: #d6d3d1;
  border-radius: 3px;
}
.chat-scroll::-webkit-scrollbar-thumb:hover {
  background: #a8a29e;
}
</style>
