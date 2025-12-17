<script lang="ts">
import { defineComponent } from 'vue'
import type { PropType } from 'vue'

export default defineComponent({
  name: 'ConfirmModal',
  props: {
    visible: { type: Boolean, required: true },
    title: { type: String, required: false, default: '提示' },
    message: { type: String, required: false, default: '' },
    confirmText: { type: String, required: false, default: '确认' },
    cancelText: { type: String, required: false, default: '取消' },
    onConfirm: { type: Function as PropType<() => void>, required: false },
    onCancel: { type: Function as PropType<() => void>, required: false },
  },
  setup(props) {
    function handleConfirm() {
      props.onConfirm && props.onConfirm()
    }
    function handleCancel() {
      props.onCancel && props.onCancel()
    }
    return { handleConfirm, handleCancel }
  }
})
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
    <div class="w-80 rounded-xl bg-[#fdf6e3] p-6 shadow-2xl border-2 border-amber-700/20">
      <h3 class="mb-4 text-xl font-bold text-amber-900 border-b border-amber-200 pb-2">{{ title }}</h3>
      <p class="mb-6 text-base text-amber-800/80 leading-relaxed">{{ message }}</p>
      <div class="flex justify-end gap-3">
        <button class="rounded-lg border border-amber-600 text-amber-700 px-4 py-2 hover:bg-amber-100 transition font-bold text-sm" @click="handleCancel">{{ cancelText }}</button>
        <button class="rounded-lg bg-amber-700 px-4 py-2 text-white hover:bg-amber-800 transition shadow-md font-bold text-sm" @click="handleConfirm">{{ confirmText }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* small adjustments to match existing project look */
</style>
