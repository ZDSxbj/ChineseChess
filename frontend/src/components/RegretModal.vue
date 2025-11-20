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
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="w-64 rounded-lg bg-white p-6 shadow-lg">
      <h3 class="mb-4 text-xl font-bold">{{ title }}</h3>
      <p v-if="showCountdown" class="mb-4">
        {{ message }} {{ countdown }}秒后自动{{ autoAction }}
      </p>
      <p v-else class="mb-4">{{ message }}</p>
      <div v-if="showButtons" class="flex justify-around">
        <button
          class="rounded bg-green-500 px-4 py-2 text-white"
          @click="handleAccept"
        >
          同意
        </button>
        <button
          class="rounded bg-red-500 px-4 py-2 text-white"
          @click="handleReject"
        >
          拒绝
        </button>
      </div>
    </div>
  </div>
</template>
