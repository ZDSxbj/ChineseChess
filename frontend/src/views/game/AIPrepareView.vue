<script lang="ts" setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const difficulty = ref<number>(3) // 默认中等难度
const playerColor = ref<'red' | 'black'>('red') // 默认执红

const difficultyOptions = [
  { value: 1, label: '简单', description: '适合新手' },
  { value: 3, label: '中等', description: '适合进阶' },
  { value: 5, label: '困难', description: '挑战高手' },
]

function startGame() {
  console.log('Starting AI game with difficulty:', difficulty.value, 'color:', playerColor.value)
  // 将设置通过路由参数传递给AI对战页面
  router.push({
    name: 'game-ai',
    query: {
      difficulty: difficulty.value.toString(),
      color: playerColor.value,
    },
  })
}

function goBack() {
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen bg-[#fdf6e3] flex items-center justify-center p-4 relative overflow-hidden">
    <!-- 背景装饰 -->
    <div class="absolute inset-0 pointer-events-none overflow-hidden">
      <div class="absolute -top-[20%] -left-[10%] w-[70%] h-[70%] rounded-full bg-amber-200/20 blur-3xl"></div>
      <div class="absolute top-[40%] -right-[10%] w-[60%] h-[60%] rounded-full bg-orange-200/20 blur-3xl"></div>
    </div>

    <main class="relative w-full max-w-2xl bg-white/60 backdrop-blur-xl rounded-3xl shadow-2xl border border-white/50 p-6 sm:p-10 animate-fade-in">
      <!-- 标题区域 -->
      <div class="text-center mb-10">
        <div class="inline-flex items-center justify-center w-16 h-16 bg-amber-100 text-amber-800 rounded-2xl mb-4 shadow-inner">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
        </div>
        <h2 class="text-3xl font-black text-amber-900 tracking-wider">人机对战设置</h2>
        <p class="text-amber-700/60 mt-2 font-medium">配置您的专属对局体验</p>
      </div>

      <!-- 难度选择 -->
      <div class="mb-8">
        <h3 class="text-lg font-bold text-amber-900 mb-4 flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          选择难度
        </h3>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div
            v-for="option in difficultyOptions"
            :key="option.value"
            class="relative cursor-pointer group rounded-xl p-4 transition-all duration-300 border-2"
            :class="difficulty === option.value
              ? 'border-amber-500 bg-amber-50 shadow-md'
              : 'border-amber-100 bg-white/50 hover:border-amber-300 hover:bg-white'"
            @click="difficulty = option.value"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-bold text-lg" :class="difficulty === option.value ? 'text-amber-700' : 'text-amber-900/70'">
                {{ option.label }}
              </span>
              <div
                class="w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors"
                :class="difficulty === option.value
                  ? 'border-amber-500 bg-amber-500'
                  : 'border-amber-200 group-hover:border-amber-300'"
              >
                <div v-if="difficulty === option.value" class="w-2 h-2 bg-white rounded-full" />
              </div>
            </div>
            <p class="text-xs font-medium" :class="difficulty === option.value ? 'text-amber-600/80' : 'text-amber-800/40'">
              {{ option.description }}
            </p>
          </div>
        </div>
      </div>

      <!-- 颜色选择 -->
      <div class="mb-10">
        <h3 class="text-lg font-bold text-amber-900 mb-4 flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
          </svg>
          选择执子
        </h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <!-- 红方 -->
          <div
            class="relative cursor-pointer rounded-xl p-4 flex items-center gap-4 border-2 transition-all duration-300 group"
            :class="playerColor === 'red'
              ? 'border-red-500 bg-red-50/50 shadow-md'
              : 'border-amber-100 bg-white/50 hover:border-red-200 hover:bg-white'"
            @click="playerColor = 'red'"
          >
            <div class="w-12 h-12 rounded-full bg-gradient-to-br from-red-100 to-red-200 flex items-center justify-center text-xl shadow-inner border border-red-200">
              <div class="w-8 h-8 rounded-full bg-red-500 shadow-lg border-2 border-red-300"></div>
            </div>
            <div class="flex-1">
              <div class="flex items-center justify-between">
                <span class="font-bold text-lg text-gray-800">执红先行</span>
                <div
                  class="w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors"
                  :class="playerColor === 'red'
                    ? 'border-red-500 bg-red-500'
                    : 'border-gray-200 group-hover:border-red-300'"
                >
                  <div v-if="playerColor === 'red'" class="w-2 h-2 bg-white rounded-full" />
                </div>
              </div>
              <p class="text-xs text-gray-500 mt-1">红方先手，主动出击</p>
            </div>
          </div>

          <!-- 黑方 -->
          <div
            class="relative cursor-pointer rounded-xl p-4 flex items-center gap-4 border-2 transition-all duration-300 group"
            :class="playerColor === 'black'
              ? 'border-gray-700 bg-gray-100 shadow-md'
              : 'border-amber-100 bg-white/50 hover:border-gray-300 hover:bg-white'"
            @click="playerColor = 'black'"
          >
            <div class="w-12 h-12 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xl shadow-inner border border-gray-200">
              <div class="w-8 h-8 rounded-full bg-gray-800 shadow-lg border-2 border-gray-600"></div>
            </div>
            <div class="flex-1">
              <div class="flex items-center justify-between">
                <span class="font-bold text-lg text-gray-800">执黑后行</span>
                <div
                  class="w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors"
                  :class="playerColor === 'black'
                    ? 'border-gray-800 bg-gray-800'
                    : 'border-gray-200 group-hover:border-gray-400'"
                >
                  <div v-if="playerColor === 'black'" class="w-2 h-2 bg-white rounded-full" />
                </div>
              </div>
              <p class="text-xs text-gray-500 mt-1">黑方后手，稳扎稳打</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="flex flex-col-reverse sm:flex-row gap-4 pt-4 border-t border-amber-100">
        <button
          class="flex-1 py-3.5 px-6 rounded-xl font-bold text-amber-800 bg-amber-100 hover:bg-amber-200 transition-all duration-300 flex items-center justify-center gap-2"
          @click="goBack"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          返回首页
        </button>
        <button
          class="flex-[2] py-3.5 px-6 rounded-xl font-bold text-white bg-gradient-to-r from-amber-600 to-amber-700 hover:from-amber-500 hover:to-amber-600 shadow-lg hover:shadow-xl hover:-translate-y-0.5 transition-all duration-300 flex items-center justify-center gap-2"
          @click="startGame"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          开始对弈
        </button>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* 移除旧的样式，使用 Tailwind 类 */
</style>
