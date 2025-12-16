<template>
  <div class="flex min-h-[calc(100vh-64px)] bg-[#fdf6e3] animate-fade-in">
    <div class="w-64 bg-amber-50/50 border-r border-amber-200 shadow-sm flex flex-col py-8">
      <h2 class="text-2xl font-black text-amber-900 text-center mb-8 tracking-widest flex items-center justify-center gap-2">
        <div class="i-carbon-user-avatar text-3xl"></div>
        个人中心
      </h2>
      <ul class="space-y-2 px-4 list-none">
        <li>
          <RouterLink to="/profile/info" class="block px-6 py-3 rounded-xl text-amber-800/70 hover:bg-amber-100 hover:text-amber-900 transition-all font-bold flex items-center gap-3" active-class="!bg-gradient-to-r !from-amber-600 !to-amber-700 !text-white shadow-md">
            <div class="i-carbon-id-management text-xl"></div>
            个人信息
          </RouterLink>
        </li>
        <li>
          <RouterLink to="records" class="block px-6 py-3 rounded-xl text-amber-800/70 hover:bg-amber-100 hover:text-amber-900 transition-all font-bold flex items-center gap-3" active-class="!bg-gradient-to-r !from-amber-600 !to-amber-700 !text-white shadow-md">
            <div class="i-carbon-catalog text-xl"></div>
            行棋记录
          </RouterLink>
        </li>
      </ul>
    </div>
    <div class="flex-1 p-8 overflow-y-auto custom-scrollbar">
      <!-- 这里是子路由的出口，点击侧边栏链接后，对应的组件会在这里渲染 -->
      <RouterView v-slot="{ Component }">
        <transition name="slide-up" mode="out-in">
          <component :is="Component" />
        </transition>
      </RouterView>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router';
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.8s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.slide-up-enter-from {
  opacity: 0;
  transform: translateY(100px);
}

.slide-up-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}
</style>
