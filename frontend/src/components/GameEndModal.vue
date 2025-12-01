<script lang="ts">
import type { PropType } from 'vue'
import { defineComponent, ref, watch } from 'vue'

export default defineComponent({
  name: 'GameEndModal',
  props: {
    visible: { type: Boolean, required: false },
    result: { type: String as unknown as () => 'win' | 'lose' | 'draw' | null, required: false },
    onReview: { type: Function as PropType<() => void>, required: false },
    onQuit: { type: Function as PropType<() => void>, required: false },
  },
  emits: ['close'],
  setup(props, { emit }) {
    const title = ref('')
    const message = ref('')

    watch(
      () => props.visible,
      (visible) => {
        if (!visible) {
          return
        }
        if (props.result === 'win') {
          title.value = '恭喜你，你赢了！'
          message.value = '本局对局已经结束，你可以选择复盘或退出。'
        }
        else if (props.result === 'lose') {
          title.value = '很遗憾，你输了'
          message.value = '对局结束，是否进入复盘查看过程？'
        }
        else if (props.result === 'draw') {
          title.value = '对局结束，本局和棋'
          message.value = '本局已和棋，可以复盘查看。'
        }
        else {
          title.value = '对局结束'
          message.value = '对局已结束，您可以选择复盘或退出。'
        }
      },
      { immediate: true },
    )

    function handleReview() {
      if (props.onReview) {
        props.onReview()
      }
      emit('close')
    }

    function handleQuit() {
      if (props.onQuit) {
        props.onQuit()
      }
      emit('close')
    }

    return {
      title,
      message,
      handleReview,
      handleQuit,
    }
  },
})
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="w-80 rounded-lg bg-white p-6 shadow-lg">
      <h3 class="mb-4 text-xl font-bold">{{ title }}</h3>
      <p class="mb-6">{{ message }}</p>
      <div class="flex justify-around">
        <button class="rounded bg-gray-600 px-4 py-2 text-white" @click="handleQuit">退出</button>
        <button class="rounded bg-blue-600 px-4 py-2 text-white" @click="handleReview">复盘</button>
      </div>
    </div>
  </div>
</template>
