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
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="w-80 rounded-lg bg-white p-6 shadow-lg">
      <h3 class="mb-4 text-xl font-bold">{{ title }}</h3>
      <p class="mb-6 text-sm text-gray-700">{{ message }}</p>
      <div class="flex justify-around">
        <button class="rounded bg-gray-600 px-4 py-2 text-white" @click="handleCancel">{{ cancelText }}</button>
        <button class="rounded bg-blue-600 px-4 py-2 text-white" @click="handleConfirm">{{ confirmText }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* small adjustments to match existing project look */
</style>
