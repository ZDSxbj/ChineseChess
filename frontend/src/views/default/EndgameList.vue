<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { endgameScenarios } from '@/data/endgameData'
import { useUserStore } from '@/store/useStore'

const router = useRouter()
const userStore = useUserStore()

const getProgressKey = () => {
  const userId = userStore.userInfo?.id || 'guest'
  return 'endgame-progress-' + userId
}

interface ScenarioProgress {
  result?: 'win' | 'lose'
  attempts: number
  bestSteps?: number
}
const progress = ref<Record<string, ScenarioProgress>>({})

function loadProgress() {
  try {
    const raw = localStorage.getItem(getProgressKey())
    if (!raw) return
    progress.value = JSON.parse(raw)
  } catch (e) {
    console.warn('读取残局进度失败', e)
  }
}

function goPlay(id: string) {
  router.push({ path: '/game/endgame', query: { id } })
}

function getDifficultyClass(difficulty: string) {
  switch (difficulty) {
    case '高级':
      return 'bg-red-100 text-red-700'
    case '中级':
      return 'bg-blue-100 text-blue-700'
    case '初级':
      return 'bg-green-100 text-green-700'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

onMounted(() => loadProgress())
</script>

<template>
  <div class="mx-auto flex h-full w-full max-w-6xl flex-col gap-6 p-6 animate-fade-in-right">
    <header class="flex items-center justify-between bg-amber-100/50 p-4 rounded-xl border border-amber-200 shadow-sm">
      <div class="flex flex-col">
        <span class="text-2xl font-black text-amber-900 flex items-center gap-2">
           <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/></svg>
           残局挑战
        </span>
        <span class="text-sm text-amber-700 mt-1">选择一个经典残局开始挑战</span>
      </div>
      <div class="rounded-full bg-amber-900 px-4 py-2 text-sm font-semibold text-amber-50 shadow-md">对手：AI 困难模式</div>
    </header>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <button
        v-for="item in endgameScenarios"
        :key="item.id"
        class="group relative flex flex-col gap-3 overflow-hidden rounded-2xl border border-amber-100 bg-gradient-to-br from-white to-amber-50 p-5 text-left shadow-sm transition-all hover:border-amber-300 hover:shadow-md hover:-translate-y-1"
        @click="goPlay(item.id)"
      >
        <!-- 第一行：标题 + 难度 -->
        <div class="flex w-full items-center justify-between">
          <span class="text-xl font-black text-gray-900">{{ item.name }}</span>
          <span class="rounded-full px-3 py-1 text-xs font-bold" :class="getDifficultyClass(item.difficulty)">
            {{ item.difficulty }}
          </span>
        </div>

        <!-- 第二行：尝试次数 -->
        <div class="text-sm text-gray-600 font-medium">
          尝试次数：{{ progress[item.id]?.attempts || 0 }}
        </div>

        <!-- 第三行：标签 + 状态 -->
        <div class="mt-2 flex flex-wrap gap-2">
          <!-- Tags -->
          <span
            v-for="tag in item.tags"
            :key="tag"
            class="rounded-lg bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600"
          >
            {{ tag }}
          </span>

          <!-- Status -->
          <span
            v-if="progress[item.id]?.result === 'win'"
            class="rounded-lg bg-green-100 px-3 py-1 text-xs font-bold text-green-700"
          >已通关</span>
          <span
            v-else
            class="rounded-lg bg-orange-100 px-3 py-1 text-xs font-bold text-orange-700"
          >待攻克</span>
        </div>

        <!-- 装饰条 -->
        <div class="absolute left-0 top-0 h-full w-1 bg-amber-500 opacity-0 transition-opacity group-hover:opacity-100"></div>
      </button>
    </div>
  </div>
</template>
