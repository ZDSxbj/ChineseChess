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
  <main class="mx-a h-full w-9/10 bg-gray-4 p-4 sm:w-3/5 flex flex-col">
    <h2 class="mx-a block w-fit text-3xl font-bold mb-6 text-gray-800">
      人机对战设置
    </h2>

    <!-- 难度选择 -->
    <div class="bg-white rounded-xl p-6 shadow-lg mb-6">
      <h3 class="text-xl font-semibold mb-4 text-gray-700 flex items-center">
        <span class="mr-2">🎯</span>
        选择难度
      </h3>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div
          v-for="option in difficultyOptions"
          :key="option.value"
          class="relative cursor-pointer border-2 rounded-lg p-4 transition-all duration-200"
          :class="difficulty === option.value 
            ? 'border-blue-500 bg-blue-50' 
            : 'border-gray-300 bg-white hover:border-blue-300 hover:bg-blue-50'"
          @click="difficulty = option.value"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="text-lg font-bold" :class="difficulty === option.value ? 'text-blue-600' : 'text-gray-800'">
              {{ option.label }}
            </span>
            <div
              class="w-5 h-5 rounded-full border-2 flex items-center justify-center"
              :class="difficulty === option.value 
                ? 'border-blue-500 bg-blue-500' 
                : 'border-gray-400'"
            >
              <div v-if="difficulty === option.value" class="w-2 h-2 bg-white rounded-full" />
            </div>
          </div>
          <p class="text-sm text-gray-600">{{ option.description }}</p>
        </div>
      </div>
    </div>

    <!-- 颜色选择 -->
    <div class="bg-white rounded-xl p-6 shadow-lg mb-6">
      <h3 class="text-xl font-semibold mb-4 text-gray-700 flex items-center">
        <span class="mr-2">🎨</span>
        选择棋子颜色
      </h3>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div
          class="relative cursor-pointer border-2 rounded-lg p-6 transition-all duration-200 flex flex-col items-center"
          :class="playerColor === 'red' 
            ? 'border-red-500 bg-red-50' 
            : 'border-gray-300 bg-white hover:border-red-300 hover:bg-red-50'"
          @click="playerColor = 'red'"
        >
          <div class="text-4xl mb-2">🔴</div>
          <span class="text-lg font-bold" :class="playerColor === 'red' ? 'text-red-600' : 'text-gray-800'">
            执红先行
          </span>
          <p class="text-sm text-gray-600 mt-2 text-center">红方先手，主动出击</p>
          <div
            class="absolute top-4 right-4 w-5 h-5 rounded-full border-2 flex items-center justify-center"
            :class="playerColor === 'red' 
              ? 'border-red-500 bg-red-500' 
              : 'border-gray-400'"
          >
            <div v-if="playerColor === 'red'" class="w-2 h-2 bg-white rounded-full" />
          </div>
        </div>

        <div
          class="relative cursor-pointer border-2 rounded-lg p-6 transition-all duration-200 flex flex-col items-center"
          :class="playerColor === 'black' 
            ? 'border-gray-800 bg-gray-100' 
            : 'border-gray-300 bg-white hover:border-gray-400 hover:bg-gray-100'"
          @click="playerColor = 'black'"
        >
          <div class="text-4xl mb-2">⚫</div>
          <span class="text-lg font-bold" :class="playerColor === 'black' ? 'text-gray-800' : 'text-gray-800'">
            执黑后行
          </span>
          <p class="text-sm text-gray-600 mt-2 text-center">黑方后手，稳扎稳打</p>
          <div
            class="absolute top-4 right-4 w-5 h-5 rounded-full border-2 flex items-center justify-center"
            :class="playerColor === 'black' 
              ? 'border-gray-800 bg-gray-800' 
              : 'border-gray-400'"
          >
            <div v-if="playerColor === 'black'" class="w-2 h-2 bg-white rounded-full" />
          </div>
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="mt-auto flex flex-col sm:flex-row gap-4 justify-center pb-4">
      <button
        class="border-0 rounded-2xl bg-green-500 text-white p-4 px-8 transition-all duration-200 text-xl font-semibold shadow-lg"
        hover="bg-green-600 shadow-xl transform scale-105"
        @click="startGame"
      >
        🎮 开始游戏
      </button>
      <button
        class="border-0 rounded-2xl bg-gray-500 text-white p-4 px-8 transition-all duration-200 text-xl font-semibold shadow-lg"
        hover="bg-gray-600 shadow-xl transform scale-105"
        @click="goBack"
      >
        🏠 返回首页
      </button>
    </div>
  </main>
</template>

<style scoped>
button:hover {
  transform: scale(1.05);
}
</style>
