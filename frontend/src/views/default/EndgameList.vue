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
  return `endgame-progress-${userId}`
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

onMounted(() => loadProgress())
</script>

<template>
  <div class="mx-auto flex h-full w-full max-w-6xl flex-col gap-4 p-4">
    <header class="flex items-center justify-between">
      <div class="flex flex-col">
        <span class="text-xl font-bold">残局挑战</span>
        <span class="text-sm text-gray-500">选择一个经典残局开始挑战</span>
      </div>
      <div class="rounded-full bg-gray-900 px-4 py-2 text-sm font-semibold text-white">对手：AI 困难模式</div>
    </header>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <button
        v-for="item in endgameScenarios"
        :key="item.id"
        class="relative flex h-full flex-col rounded-2xl border border-gray-200 p-4 text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-md overflow-hidden"
        :class="item.id === 'qi-xing-ju-hui' ? 'bg-cover bg-center' : 'bg-gradient-to-br from-amber-50 to-white'"
        :style="item.id === 'qi-xing-ju-hui' ? 'background-image: linear-gradient(to bottom, rgba(255,248,240,0.85), rgba(255,255,255,0.9)), url(/images/endgame_qixingjuhui.png);' : ''"
        @click="goPlay(item.id)"
      >
        <div class="flex items-center justify-between relative z-10">
          <span class="text-lg font-semibold">{{ item.name }}</span>
          <span
            class="rounded-full px-3 py-1 text-xs font-semibold"
            :class="{
              'bg-green-100 text-green-700': item.difficulty === '初级',
              'bg-blue-100 text-blue-700': item.difficulty === '中级',
              'bg-red-100 text-red-700': item.difficulty === '高级',
            }"
          >
            {{ item.difficulty }}
          </span>
        </div>
        <p v-if="item.goal" class="mt-1 text-sm text-gray-700 relative z-10">{{ item.goal }}</p>
        <div class="mt-2 flex flex-col gap-2 relative z-10">
          <div class="flex items-center justify-between text-xs text-gray-600">
            <span>尝试次数：<span class="font-semibold text-gray-800">{{ progress[item.id]?.attempts || 0 }}</span></span>
            <span v-if="progress[item.id]?.bestSteps">最少步数：<span class="font-semibold text-blue-600">{{ progress[item.id].bestSteps }}</span></span>
          </div>
          <div class="flex flex-wrap gap-2 text-xs text-gray-600">
            <span v-for="tag in item.tags" :key="tag" class="rounded-full bg-gray-100 px-2 py-0.5">{{ tag }}</span>
            <span v-if="progress[item.id]?.result === 'win'" class="rounded-full bg-emerald-100 px-2 py-0.5 text-emerald-700">已功克</span>
            <span v-else class="rounded-full bg-orange-100 px-2 py-0.5 text-orange-700">待攻克</span>
          </div>
        </div>
      </button>
    </div>
  </div>
</template>
