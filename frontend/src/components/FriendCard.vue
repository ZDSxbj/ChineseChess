<script lang="ts" setup>
const props = defineProps<{
  friend: {
    unreadCount?: number
    challengePending?: boolean
    id: number
    name: string
    avatar: string
    online: boolean
    gender: string
    exp: number
    totalGames: number
    winRate: number
  }
  selected?: boolean
}>()

const emit = defineEmits(['select', 'context'])

const _base = import.meta.env.BASE_URL || '/'
const base = _base.endsWith('/') ? _base : `${_base}/`
const defaultAvatar = `${base}images/default_avatar.png`

function onSelect() {
  emit('select', props.friend.id)
}

function onContext(e: MouseEvent) {
  e.preventDefault()
  emit('context', { id: props.friend.id, x: e.clientX, y: e.clientY })
}

function handleImageError(e: Event) {
  const target = e.target as HTMLImageElement
  // 防止死循环
  // 如果浏览器将 src 解析成绝对 URL，直接比较字符串可能失败
  // 使用 endsWith 校验避免死循环（也能覆盖绝对/相对路径差异）
  if (target.src.endsWith('images/default_avatar.png'))
    return
  target.src = defaultAvatar
}
</script>

<template>
  <div
    @click="onSelect"
    @contextmenu="onContext"
    class="flex items-center gap-4 p-4 cursor-pointer transition-colors duration-200 border-b border-gray-100"
    :class="selected ? 'bg-blue-50 border-l-4 border-l-blue-500' : 'hover:bg-gray-50 bg-white'"
  >
    <img
      :src="friend.avatar || defaultAvatar"
      @error="handleImageError"
      alt="avatar"
      class="w-14 h-14 rounded-full object-cover border border-gray-200 shadow-sm"
    />
    <div class="flex-1 min-w-0">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h3 class="text-base font-semibold text-gray-800 truncate">{{ friend.name }}</h3>
          <span v-if="friend.unreadCount && friend.unreadCount > 0" class="ml-2 inline-flex items-center justify-center bg-red-500 text-white text-xs font-semibold rounded-full px-2 py-0.5">{{ friend.unreadCount > 99 ? '99+' : friend.unreadCount }}</span>
          <span v-if="friend.challengePending" class="ml-2 inline-flex items-center justify-center bg-yellow-400 text-gray-900 text-xs font-semibold rounded-full px-2 py-0.5">挑战</span>
          <!-- 性别图标 -->
          <div v-if="friend.gender === '男'" class="i-carbon-gender-male text-blue-500 text-lg" title="男"></div>
          <div v-else-if="friend.gender === '女'" class="i-carbon-gender-female text-pink-500 text-lg" title="女"></div>
          <div v-else-if="friend.gender === '其他'" class="i-carbon-help text-gray-400 text-lg" title="其他"></div>
        </div>
        <span v-if="friend.online" class="w-2.5 h-2.5 bg-green-500 rounded-full shadow-sm" title="在线"></span>
        <span v-else class="w-2.5 h-2.5 bg-gray-300 rounded-full" title="离线"></span>
      </div>

      <div class="flex items-center gap-3 mt-1.5 text-sm text-gray-500">
        <span>场次: {{ friend.totalGames }}</span>
        <span>胜率: {{ (friend.winRate * 100).toFixed(1) }}%</span>
      </div>
      <div class="text-xs text-gray-400 mt-0.5">经验: {{ friend.exp }}</div>
    </div>
  </div>
</template>
