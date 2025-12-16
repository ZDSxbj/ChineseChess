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
    hideReview: { type: Boolean, default: false },
    messageOverride: { type: String, default: '' },
  },
  emits: ['close'],
  setup(props, { emit }) {
    const title = ref('')
    const message = ref('')

    watch(
      () => [props.visible, props.result, props.messageOverride],
      ([visible]) => {
        if (!visible) {
          return
        }
        if (props.result === 'win') {
          title.value = '恭喜你，你赢了！'
          message.value = props.messageOverride || '本局对局已经结束，你可以选择复盘或退出。'
        }
        else if (props.result === 'lose') {
          title.value = '很遗憾，你输了'
          message.value = props.messageOverride || '对局结束，是否进入复盘查看过程？'
        }
        else if (props.result === 'draw') {
          title.value = '对局结束，本局和棋'
          message.value = props.messageOverride || '本局已和棋，可以复盘查看。'
        }
        else {
          title.value = '对局结束'
          message.value = props.messageOverride || '对局已结束，您可以选择复盘或退出。'
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
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
    <div class="w-96 rounded-2xl bg-[#fdf6e3] p-8 shadow-2xl border-4 border-amber-200">
      <div class="text-center mb-6">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-amber-100 mb-4 shadow-inner">
          <span class="text-4xl" v-if="result === 'win'">🏆</span>
          <span class="text-4xl" v-else-if="result === 'lose'">💔</span>
          <span class="text-4xl" v-else-if="result === 'draw'">🤝</span>
          <span class="text-4xl" v-else>🏁</span>
        </div>
        <h3 class="text-2xl font-black text-amber-900">{{ title }}</h3>
      </div>

      <p class="mb-8 text-center text-amber-800/80 font-medium leading-relaxed">{{ message }}</p>

      <div class="flex flex-col gap-3">
        <button v-if="!hideReview" class="w-full rounded-xl bg-gradient-to-r from-amber-600 to-amber-700 text-white py-3 font-bold shadow-lg hover:shadow-xl hover:-translate-y-0.5 transition-all duration-200 flex items-center justify-center gap-2" @click="handleReview">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          复盘对局
        </button>
        <button class="w-full rounded-xl border-2 border-amber-300 text-amber-800 py-3 font-bold hover:bg-amber-100 transition-colors flex items-center justify-center gap-2" @click="handleQuit">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          退出游戏
        </button>
      </div>
    </div>
  </div>
</template>
