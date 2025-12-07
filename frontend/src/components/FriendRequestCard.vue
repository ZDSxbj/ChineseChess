<script lang="ts" setup>
import { defineProps, defineEmits } from 'vue'

interface Req {
  id: number
  senderId: number
  senderName: string
  senderAvatar?: string
  content?: string
  createdAt?: number
  unreadCount?: number
  gender?: string
  exp?: number
  totalGames?: number
  winRate?: number
}

const props = defineProps<{ request: Req, selected?: boolean }>()
const emit = defineEmits<{
  (e: 'accept', id: number): void
  (e: 'reject', id: number): void
}>()

const _base = import.meta.env.BASE_URL || '/'
const base = _base.endsWith('/') ? _base : `${_base}/`
const defaultAvatar = `${base}images/default_avatar.png`

function onAccept() {
  emit('accept', props.request.id)
}
function onReject() {
  emit('reject', props.request.id)
}

function handleImageError(e: Event) {
  const target = e.target as HTMLImageElement
  if (target.src.endsWith('images/default_avatar.png')) return
  target.src = defaultAvatar
}
</script>

<template>
  <div class="flex items-start gap-4 p-4 border-b border-gray-100 transition-colors duration-150 rounded-lg bg-gradient-to-r from-yellow-50 to-white">
    <img
      :src="props.request.senderAvatar || defaultAvatar"
      @error="handleImageError"
      alt="avatar"
      class="w-14 h-14 rounded-full object-cover border border-gray-200 shadow-sm"
    />
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h3 class="text-base font-semibold text-gray-800 truncate">{{ props.request.senderName }}</h3>
          <span v-if="props.request.unreadCount && props.request.unreadCount > 0" class="ml-2 inline-flex items-center justify-center bg-red-500 text-white text-xs font-semibold rounded-full px-2 py-0.5">{{ props.request.unreadCount > 99 ? '99+' : props.request.unreadCount }}</span>
          <div v-if="props.request.gender === '男'" class="i-carbon-gender-male text-blue-500 text-lg" title="男"></div>
          <div v-else-if="props.request.gender === '女'" class="i-carbon-gender-female text-pink-500 text-lg" title="女"></div>
        </div>
        <div class="text-xs text-gray-400">{{ props.request.createdAt ? new Date(props.request.createdAt * 1000).toLocaleString() : '' }}</div>
      </div>

      <div class="flex items-center gap-3 mt-1.5 text-sm text-gray-500">
        <span>场次: {{ props.request.totalGames || 0 }}</span>
        <span>胜率: {{ ((props.request.winRate || 0) * 100).toFixed(1) }}%</span>
      </div>
      <div class="text-xs text-gray-400 mt-0.5">经验: {{ props.request.exp || 0 }}</div>

      <div class="mt-3 p-3 bg-white border border-dashed border-yellow-100 rounded-md text-sm text-gray-700 whitespace-pre-wrap">{{ props.request.content }}</div>

      <div class="flex items-center gap-2 mt-3">
        <button @click="onAccept" class="px-3 py-1 text-xs text-white bg-emerald-500 rounded-full hover:bg-emerald-600">同意</button>
        <button @click="onReject" class="px-3 py-1 text-xs text-gray-700 border border-gray-200 rounded-full hover:bg-gray-50">拒绝</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
