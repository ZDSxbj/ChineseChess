<script lang="ts" setup>
import { Icon } from '@iconify/vue/dist/iconify.js'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/store/useStore'
import { useSocialStore } from '@/store/useSocialStore'
import { onMounted } from 'vue'

const userStore = useUserStore()
const { userInfo } = storeToRefs(userStore)
const socialStore = useSocialStore()
const { totalUnread } = storeToRefs(socialStore)

onMounted(() => {
  // 初始化 social store，使其在全局订阅消息并计算初始未读数
  try { socialStore.init() } catch { /* ignore */ }
})
</script>

<template>
  <div class="h-full w-full flex flex-col bg-[#fdf6e3]">
    <header class="bg-amber-900 text-amber-50 py-3 px-4 shadow-lg sticky top-0 z-50 flex justify-between items-center">
      <div class="flex items-center gap-2 cursor-pointer" @click="$router.push('/')">
        <div class="w-8 h-8 bg-amber-100 text-amber-900 rounded flex items-center justify-center font-black border-2 border-amber-600">象</div>
        <span class="font-bold text-lg tracking-widest">弈苑</span>
      </div>

      <nav class="hidden sm:flex gap-4 text-sm font-medium items-center">
          <router-link to="/" class="hover:text-amber-200 transition px-3 py-2 rounded-md" exact-active-class="text-amber-200 bg-amber-800">
            首页
          </router-link>

          <router-link to="/social" class="hover:text-amber-200 transition px-3 py-2 rounded-md flex items-center gap-1" active-class="text-amber-200 bg-amber-800">
            <span>社交</span>
            <span v-if="totalUnread && totalUnread > 0" class="inline-flex items-center justify-center bg-red-500 text-white text-[10px] font-bold rounded-full px-1.5 py-0.5 min-w-[1.2em]">{{ totalUnread > 99 ? '99+' : totalUnread }}</span>
          </router-link>

          <router-link to="/endgame" class="hover:text-amber-200 transition px-3 py-2 rounded-md" active-class="text-amber-200 bg-amber-800">
            残局挑战
          </router-link>

          <router-link
            v-if="userInfo?.name"
            to="/profile/info"
            class="hover:text-amber-200 transition px-3 py-2 rounded-md flex items-center gap-2"
            active-class="text-amber-200 bg-amber-800"
          >
            <div class="w-6 h-6 rounded-full bg-amber-200 text-amber-900 flex items-center justify-center text-xs font-bold">
               {{ userInfo.name.charAt(0).toUpperCase() }}
            </div>
            {{ userInfo.name }}
          </router-link>

          <RouterLink
            v-else
            to="/auth/login"
            class="bg-amber-100 text-amber-900 px-4 py-2 rounded-md font-bold hover:bg-white transition shadow-sm"
          >
            登录
          </RouterLink>
      </nav>

      <!-- Mobile Menu -->
      <nav class="group relative sm:hidden">
        <Icon width="28" height="28" icon="ion:navicon-round" class="text-amber-100" />
        <div
          class="absolute right-0 top-full mt-2 hidden flex-col gap-1 rounded-xl bg-amber-50 p-2 shadow-xl border border-amber-200 min-w-[150px]"
          group-focus="flex"
          group-hover="flex"
          tabindex="0"
        >
          <router-link to="/" class="text-amber-900 hover:bg-amber-100 px-4 py-3 rounded-lg transition" active-class="bg-amber-200 font-bold">
            首页
          </router-link>

          <router-link to="/social" class="text-amber-900 hover:bg-amber-100 px-4 py-3 rounded-lg transition flex justify-between items-center" active-class="bg-amber-200 font-bold">
            <span>社交</span>
            <span v-if="totalUnread && totalUnread > 0" class="bg-red-500 text-white text-xs font-bold rounded-full px-2 py-0.5">{{ totalUnread > 99 ? '99+' : totalUnread }}</span>
          </router-link>

          <router-link to="/endgame" class="text-amber-900 hover:bg-amber-100 px-4 py-3 rounded-lg transition" active-class="bg-amber-200 font-bold">
            残局挑战
          </router-link>

          <div class="h-px bg-amber-200 my-1"></div>

          <router-link
            v-if="userInfo?.name"
            to="/profile/info"
            class="text-amber-900 hover:bg-amber-100 px-4 py-3 rounded-lg transition font-bold"
            active-class="bg-amber-200"
          >
            {{ userInfo.name }}
          </router-link>

          <RouterLink
            v-else
            to="/auth/login"
            class="text-amber-900 hover:bg-amber-100 px-4 py-3 rounded-lg transition font-bold"
          >
            登录
          </RouterLink>
        </div>
      </nav>
    </header>
    <main class="flex-1 overflow-hidden flex flex-col relative">
      <RouterView />
    </main>
  </div>
</template>
