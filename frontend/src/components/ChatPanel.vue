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
  <div class="w-full flex flex-col bg-white rounded-lg shadow-sm" :class="[isMinimized ? 'h-12' : 'h-chat']">
    <div class="px-4 py-2 border-b text-gray-700 font-medium flex justify-between items-center">
      <span>聊天</span>
      <button
        class="w-6 h-6 flex items-center justify-center text-gray-500 hover:text-gray-700 focus:outline-none"
        @click="isMinimized = !isMinimized"
      >
        <svg v-if="isMinimized" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
    </div>

    <div v-show="!isMinimized" ref="chatScrollRef" class="flex-1 p-4 overflow-y-auto chat-scroll">
      <div v-for="(msg, index) in messages" :key="index" class="mb-4">
        <div class="message-row" :class="[msg.sender === '我' ? 'self' : 'other']">
          <div class="bubble" :class="[msg.sender === '我' ? 'bubble--self' : 'bubble--other']">
            <div class="bubble-content">{{ msg.content }}</div>
            <div class="bubble-time">{{ msg.time }}</div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="!isMinimized" class="px-3 py-3 border-t bg-gray-50 chat-input">
      <div class="flex items-center space-x-3">
        <input
          v-model="inputMessage"
          type="text"
          :maxlength="maxLen"
          class="flex-1 px-3 py-3 border rounded text-sm outline-none focus:ring-2 focus:ring-blue-300"
          placeholder="请输入信息..."
          @keyup.enter="sendMessage"
        />
        <div class="text-xs text-gray-500">{{ inputMessage.length }}/{{ maxLen }}</div>
        <button
          class="ml-2 px-4 py-2 bg-gray-700 text-white rounded hover:bg-gray-800 text-sm"
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
  /* 占满父容器高度，避免随着消息增长改变自身高度 */
  height: 100%;
}
.chat-scroll {
  /* 消息区占据剩余空间并可滚动；min-height:0 允许在 flex 容器内正确滚动 */
  flex: 1 1 0%;
  min-height: 0;
  overflow-y: auto;
}

.chat-input {
  /* 输入区固定高度，避免被消息区挤压 */
  flex: 0 0 64px;
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
  max-width: 70%;
  padding: 10px 12px;
  border-radius: 12px;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.02);
}
.bubble--self {
  background: #007bff;
  color: white;
  border-bottom-right-radius: 4px;
}
.bubble--other {
  background: #f1f5f9;
  color: #111827;
  border-bottom-left-radius: 4px;
}

.bubble-content {
  white-space: pre-wrap;
  word-break: break-word;
}
.bubble-time {
  font-size: 11px;
  color: rgba(0, 0, 0, 0.45);
  margin-top: 6px;
}

.chat-scroll::-webkit-scrollbar {
  width: 6px;
}
.chat-scroll::-webkit-scrollbar-track {
  background: #f7fafc;
}
.chat-scroll::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 3px;
}
</style>
