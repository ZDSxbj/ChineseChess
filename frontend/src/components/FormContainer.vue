<script setup lang="ts">
import type { CustomFormData } from '@/composables/FormExam'

import FormInput from '@/components/FormInput.vue'

defineProps({
  formName: {
    type: String,
    required: false,
    default: '提交',
  },
  disabled: {
    type: Boolean,
    required: false,
    default: false,
  },
  formData: {
    type: Array<CustomFormData>,
    // type: Array as () => CustomFormData[],
    required: true,
  },
})

defineEmits(['submitForm'])
</script>

<template>
  <form class="mx-auto border border-amber-200 p-12 shadow-2xl rounded-xl bg-[#fdf6e3]/90 backdrop-blur-sm max-w-md w-full" @submit.prevent="$emit('submitForm')">
    <FormInput
      v-for="item in formData"
      :id="item.id"
      :key="item.id"
      v-model="item.value"
      :label="item.label"
      :type="item.type || 'text'"
      :autocomplete="item.autocomplete || 'off'"
    />
    <slot />
    <button
      :disabled="disabled"
      class="mt-8 h-12 w-full flex cursor-pointer items-center justify-center border-0 rounded-xl bg-amber-700 text-lg font-bold text-white disabled:bg-gray-400 disabled:cursor-not-allowed hover:bg-amber-800 transition-all shadow-md hover:shadow-lg active:scale-95"
      type="submit"
    >
      {{ disabled ? '请填写完整信息' : formName }}
    </button>
  </form>
</template>
