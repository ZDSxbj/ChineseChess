<script lang="ts" setup>
import { ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  type: 'requesting' | 'responding'
  mode?: 'regret' | 'draw'
  onAccept?: () => void
  onReject?: () => void
}>()

const emit = defineEmits(['close'])

const title = ref('')
const message = ref('')
const countdown = ref(10)
const showCountdown = ref(false)
const showButtons = ref(false)
const autoAction = ref('')
let countdownTimer: NodeJS.Timeout | null = null

// 清除定时器（通用方法）
function clearTimer() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

// 开始倒计时（仅响应方需要）
function startCountdown() {
  clearTimer() // 先清除可能存在的定时器
  countdown.value = 10
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearTimer()
      props.onReject?.() // 超时自动拒绝
      emit('close')
    }
  }, 1000)
}

// 监听 visible 和 type 变化，动态初始化状态
watch(
  () => [props.visible, props.type],
  ([newVisible, newType]) => {
    if (newVisible) {
      // 显示时根据 type 和 mode 初始化
      const m = props.mode || 'regret'
      if (newType === 'requesting') {
        title.value = '等待回应'
        message.value = m === 'draw' ? '对方正在考虑是否同意和棋' : '对方正在考虑是否同意悔棋'
        showCountdown.value = false
        showButtons.value = false
        clearTimer() // 请求方不需要倒计时
      }
      else {
        title.value = m === 'draw' ? '和棋请求' : '悔棋请求'
        message.value = m === 'draw' ? '对方请求和棋，是否同意？' : '对方请求悔棋，是否同意？'
        showCountdown.value = true
        showButtons.value = true
        autoAction.value = '拒绝'
        startCountdown() // 响应方启动倒计时
      }
    }
    else {
      // 隐藏时清除所有状态
      clearTimer()
    }
  },
  { immediate: true }, // 初始化时立即执行
)

// 同意悔棋
function handleAccept() {
  clearTimer()
  props.onAccept?.()
  emit('close')
}

// 拒绝悔棋
function handleReject() {
  clearTimer()
  props.onReject?.()
  emit('close')
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
    <div class="w-80 rounded-2xl bg-[#fdf6e3] p-6 shadow-2xl border-4 border-amber-200">
      <h3 class="mb-4 text-xl font-black text-amber-900 border-b border-amber-200 pb-3 text-center">{{ title }}</h3>

      <div class="text-center mb-6">
        <p v-if="showCountdown" class="text-amber-800 font-medium">
          {{ message }}
          <br/>
          <span class="text-sm text-amber-600 mt-2 block font-bold bg-amber-100 py-1 px-3 rounded-full inline-block mx-auto">
            {{ countdown }}秒后自动{{ autoAction }}
          </span>
        </p>
        <p v-else class="text-amber-800 font-medium flex items-center justify-center gap-2">
          <svg class="animate-spin h-5 w-5 text-amber-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ message }}
        </p>
      </div>

      <div v-if="showButtons" class="flex justify-center gap-4">
        <button
          class="flex-1 rounded-xl bg-emerald-600 text-white px-4 py-2.5 font-bold shadow-md hover:bg-emerald-700 transition-colors flex items-center justify-center gap-1"
          @click="handleAccept"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          同意
        </button>
        <button
          class="flex-1 rounded-xl bg-red-500 text-white px-4 py-2.5 font-bold shadow-md hover:bg-red-600 transition-colors flex items-center justify-center gap-1"
          @click="handleReject"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
          拒绝
        </button>
      </div>
    </div>
  </div>
</template>
