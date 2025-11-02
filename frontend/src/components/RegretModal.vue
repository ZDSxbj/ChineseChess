<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  visible: boolean
  type: 'requesting' | 'responding'
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

function startCountdown() {
  countdown.value = 10
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countdownTimer as NodeJS.Timeout)
      props.onReject?.()
      emit('close')
    }
  }, 1000)
}

onMounted(() => {
  if (props.type === 'requesting') {
    title.value = '等待回应'
    message.value = '对方正在考虑是否同意悔棋'
    showCountdown.value = false
    showButtons.value = false
  }
  else {
    title.value = '悔棋请求'
    message.value = '对方请求悔棋，是否同意？'
    showCountdown.value = true
    showButtons.value = true
    autoAction.value = '拒绝'
    startCountdown()
  }
})

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
  }
})

function handleAccept() {
  if (countdownTimer) {
    clearInterval(countdownTimer) // 清除倒计时定时器，避免自动拒绝
  }
  props.onAccept?.() // 调用父组件传递的“同意”回调
  emit('close') // 通知父组件关闭弹窗
}

function handleReject() {
  if (countdownTimer) {
    clearInterval(countdownTimer) // 清除倒计时定时器
  }
  props.onReject?.() // 调用父组件传递的“拒绝”回调
  emit('close') // 通知父组件关闭弹窗
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
