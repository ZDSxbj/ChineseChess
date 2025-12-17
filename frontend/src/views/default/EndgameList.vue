<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { endgameScenarios } from '@/data/endgameData'
import { useUserStore } from '@/store/useStore'
import { getEndgameProgress } from '@/api/endgame/progress'

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
  const userId = userStore.userInfo?.id

  // 已登录：优先从后端取，确保不同浏览器/设备一致
  if (userId) {
    getEndgameProgress()
      .then((resp: any) => {
        const data = resp && typeof resp === 'object' && 'data' in resp ? resp.data : resp
        if (!data || !Array.isArray(data.progress)) return
        const map: Record<string, ScenarioProgress> = {}
        data.progress.forEach((item: any) => {
          const id = String(item.scenario_id)
          const attempts = Number(item.attempts) || 0
          const bestSteps = typeof item.best_steps === 'number' ? item.best_steps : undefined
          const lastResult = item.last_result === 'win' || item.last_result === 'lose' ? item.last_result : undefined
          map[id] = { attempts, bestSteps, result: lastResult }
        })
        progress.value = map
        // 同步一份到本地，用作离线缓存
        try {
          localStorage.setItem(getProgressKey(), JSON.stringify(progress.value))
        } catch {}
      })
      .catch((e: any) => {
        console.warn('从后端获取残局进度失败，回退到本地缓存', e)
        try {
          const raw = localStorage.getItem(getProgressKey())
          if (!raw) return
          progress.value = JSON.parse(raw)
        } catch (err) {
          console.warn('读取本地残局进度失败', err)
        }
      })
    return
  }

  // 未登录：仅使用本地缓存
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
  <div class="mx-auto flex h-full w-full max-w-[1400px] flex-col gap-8 p-8 animate-fade-in-right">
    <header class="flex items-center justify-between bg-amber-100/50 p-6 rounded-2xl border border-amber-200 shadow-sm">
      <div class="flex flex-col gap-1">
        <span class="text-3xl font-black text-amber-900 flex items-center gap-3">
           <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/></svg>
           残局挑战
        </span>
        <span class="text-base text-amber-700 mt-1">选择一个经典残局开始挑战</span>
      </div>
      <div class="rounded-full bg-amber-900 px-6 py-3 text-base font-semibold text-amber-50 shadow-md">对手：AI 困难模式</div>
    </header>

    <div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
      <button
        v-for="item in endgameScenarios"
        :key="item.id"
        class="group relative flex flex-col gap-4 overflow-hidden rounded-2xl border border-amber-100 bg-gradient-to-br from-white to-amber-50 p-6 text-left shadow-sm transition-all hover:border-amber-300 hover:shadow-lg hover:-translate-y-1"
        @click="goPlay(item.id)"
      >
        <!-- 第一行：标题 + 难度 -->
        <div class="flex w-full items-center justify-between">
          <span class="text-2xl font-black text-gray-900">{{ item.name }}</span>
          <span class="rounded-full px-4 py-1.5 text-sm font-bold" :class="getDifficultyClass(item.difficulty)">
            {{ item.difficulty }}
          </span>
        </div>

        <!-- 第二行：尝试次数 -->
        <div class="text-base text-gray-600 font-medium">
          尝试次数：{{ progress[item.id]?.attempts || 0 }}
        </div>

        <!-- 第三行：标签 + 状态 -->
        <div class="mt-2 flex flex-wrap gap-2">
          <!-- Tags -->
          <span
            v-for="tag in item.tags"
            :key="tag"
            class="rounded-lg bg-slate-100 px-3 py-1.5 text-sm font-bold text-slate-600"
          >
            {{ tag }}
          </span>

          <!-- Status -->
          <span
            v-if="progress[item.id]?.result === 'win'"
            class="rounded-lg bg-green-100 px-3 py-1.5 text-sm font-bold text-green-700"
          >已通关</span>
          <span
            v-else
            class="rounded-lg bg-orange-100 px-3 py-1.5 text-sm font-bold text-orange-700"
          >待攻克</span>
        </div>

        <!-- 装饰条 -->
        <div class="absolute left-0 top-0 h-full w-1.5 bg-amber-500 opacity-0 transition-opacity group-hover:opacity-100"></div>
      </button>
    </div>
  </div>
</template>
