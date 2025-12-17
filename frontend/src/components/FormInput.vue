<script setup lang="ts">
defineOptions({
  inheritAttrs: false,
})
const info = defineModel({
  type: String,
  required: true,
})
</script>

<template>
  <div class="relative mb-6 h-12 w-full">
    <input
      v-model="info"
      class="peer h-full w-full border-b-2 border-amber-200 bg-transparent text-lg text-amber-900 focus:outline-none pt-4 placeholder-transparent"
      v-bind="$attrs"
      required
      placeholder="placeholder"
    >
    <div
      class="absolute bottom-0 h-[2px] w-full scale-x-0 bg-amber-700 transition-all duration-300 ease peer-focus:scale-x-100 peer-valid:scale-x-100"
    />
    <label
      class="pointer-events-none absolute left-0 bottom-2 text-amber-700/60 transition-all duration-300 ease
             peer-focus:-translate-y-6 peer-focus:text-xs peer-focus:text-amber-700 peer-focus:font-bold
             peer-valid:-translate-y-6 peer-valid:text-xs peer-valid:text-amber-700 peer-valid:font-bold"
      :for="$attrs.id as string"
    >{{ $attrs.label }}</label>
  </div>
</template>
