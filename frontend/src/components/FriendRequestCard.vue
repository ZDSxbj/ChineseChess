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
  <div class="flex items-start gap-4 p-4 border border-amber-200/50 transition-all duration-300 rounded-xl bg-gradient-to-r from-amber-50 to-white shadow-sm hover:shadow-md hover:scale-[1.02] mb-3 mx-2 relative overflow-hidden group">
    <!-- Decorative background element -->
    <div class="absolute top-0 right-0 w-16 h-16 bg-amber-100/50 rounded-bl-full -mr-8 -mt-8 transition-transform group-hover:scale-150 duration-500"></div>

    <img
      :src="props.request.senderAvatar || defaultAvatar"
      @error="handleImageError"
      alt="avatar"
      class="w-14 h-14 rounded-full object-cover border-2 border-amber-200 shadow-sm z-10"
    />
    <div class="flex-1 min-w-0 z-10">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h3 class="text-base font-black text-amber-900 truncate">{{ props.request.senderName }}</h3>
          <span v-if="props.request.unreadCount && props.request.unreadCount > 0" class="ml-2 inline-flex items-center justify-center bg-red-600 text-white text-xs font-bold rounded-full px-2 py-0.5 shadow-sm animate-pulse">{{ props.request.unreadCount > 99 ? '99+' : props.request.unreadCount }}</span>
          <div v-if="props.request.gender === '男'" class="i-carbon-gender-male text-blue-600 text-lg" title="男"></div>
          <div v-else-if="props.request.gender === '女'" class="i-carbon-gender-female text-pink-600 text-lg" title="女"></div>
        </div>
        <div class="text-xs text-amber-800/60 font-medium">{{ props.request.createdAt ? new Date(props.request.createdAt * 1000).toLocaleString() : '' }}</div>
      </div>

      <div class="flex items-center gap-3 mt-1.5 text-sm text-amber-800/70 font-medium">
        <span>场次: {{ props.request.totalGames || 0 }}</span>
        <span>胜率: {{ (props.request.winRate || 0).toFixed(2) }}%</span>
      </div>
      <div class="text-xs text-amber-800/50 mt-0.5 font-medium">经验: {{ props.request.exp || 0 }}</div>

      <div class="mt-3 p-3 bg-white/80 border border-dashed border-amber-300 rounded-lg text-sm text-amber-900 whitespace-pre-wrap font-serif shadow-inner relative">
        <div class="absolute -top-2 -left-2 text-amber-200 text-2xl opacity-50">❝</div>
        {{ props.request.content || '请求添加好友' }}
        <div class="absolute -bottom-4 -right-1 text-amber-200 text-2xl opacity-50">❞</div>
      </div>

      <div class="flex items-center justify-end gap-3 mt-3">
        <button @click="onReject" class="px-4 py-1.5 text-xs font-bold text-amber-800 border border-amber-300 rounded-full hover:bg-amber-50 transition-colors">拒绝</button>
        <button @click="onAccept" class="px-5 py-1.5 text-xs font-bold text-white bg-gradient-to-r from-emerald-600 to-emerald-700 rounded-full hover:from-emerald-700 hover:to-emerald-800 shadow-md hover:shadow-lg transition-all active:scale-95 flex items-center gap-1">
          <div class="i-carbon-checkmark"></div>
          同意
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
