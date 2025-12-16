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
    class="relative flex items-center gap-4 p-4 cursor-pointer transition-all duration-300 mb-2 mx-2 rounded-xl shadow-sm hover:shadow-md hover:scale-[1.02] overflow-hidden group"
    :class="selected ? 'bg-gradient-to-r from-amber-200 to-amber-100 shadow-inner' : 'bg-gradient-to-r from-white to-amber-50 hover:from-amber-50 hover:to-amber-100'"
  >
    <!-- 左侧边框修饰 -->
    <div
      class="absolute left-0 top-0 bottom-0 w-1.5 transition-colors duration-300"
      :class="selected ? 'bg-amber-700' : 'bg-transparent group-hover:bg-amber-400'"
    ></div>

    <img
      :src="props.friend.avatar || defaultAvatar"
      @error="handleImageError"
      alt="avatar"
      class="w-14 h-14 rounded-full object-cover border-2 border-amber-200 shadow-sm z-10"
    />
    <div class="flex-1 min-w-0 z-10">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h3 class="text-base font-bold text-amber-900 truncate">{{ props.friend.name }}</h3>
          <span v-if="props.friend.unreadCount && props.friend.unreadCount > 0" class="ml-2 inline-flex items-center justify-center bg-red-600 text-white text-xs font-bold rounded-full px-2 py-0.5 shadow-sm animate-pulse">{{ props.friend.unreadCount > 99 ? '99+' : props.friend.unreadCount }}</span>
          <span v-if="props.friend.challengePending" class="ml-2 inline-flex items-center justify-center bg-amber-400 text-amber-900 text-xs font-bold rounded-full px-2 py-0.5 shadow-sm animate-bounce">挑战</span>
          <!-- 性别标识：使用字符避免图标依赖 -->
          <span v-if="props.friend.gender === '男'" class="text-blue-600 text-lg font-bold" title="男">♂</span>
          <span v-else-if="props.friend.gender === '女'" class="text-pink-600 text-lg font-bold" title="女">♀</span>
          <span v-else-if="props.friend.gender === '其他'" class="text-gray-500 text-lg font-bold" title="其他">？</span>
        </div>
        <span v-if="props.friend.online" class="w-2.5 h-2.5 bg-green-600 rounded-full shadow-sm border border-white animate-pulse" title="在线"></span>
        <span v-else class="w-2.5 h-2.5 bg-gray-400 rounded-full border border-white" title="离线"></span>
      </div>

      <div class="flex items-center gap-3 mt-1.5 text-sm text-amber-800/70 font-medium">
        <span>场次: {{ props.friend.totalGames }}</span>
        <span>胜率: {{ (props.friend.winRate || 0).toFixed(2) }}%</span>
      </div>
      <div class="text-xs text-amber-800/50 mt-0.5 font-medium">经验: {{ props.friend.exp }}</div>
    </div>
  </div>
</template>
