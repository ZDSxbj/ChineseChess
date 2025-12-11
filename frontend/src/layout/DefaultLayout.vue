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
  <div class="h-full w-full flex flex-col">
    <header class="w-full flex flex-row items-center justify-between p-4 shadow-2xl">
      <h1 class="w-fit">
        ChineseChess
      </h1>
      <nav class="hidden w-fit sm:block">
        <ul class="mr-6 flex flex-row gap-4">
          <router-link to="/" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            首页
          </router-link>
          
          <router-link to="/social" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            <span class="inline-flex items-center gap-2">
              <span>社交</span>
              <span v-if="totalUnread && totalUnread > 0" class="ml-2 inline-flex items-center justify-center bg-red-500 text-white text-xs font-semibold rounded-full px-2 py-0.5">{{ totalUnread > 99 ? '99+' : totalUnread }}</span>
            </span>
          </router-link>

          <!-- 修改点 1: 将 <a> 标签改为 <router-link> -->
          <router-link
            v-if="userInfo?.name"
            to="/profile/info"
            class="rounded-lg bg-[#e0e0e0] p-4 text-xl"
            hover="bg-[#b1aeae]"
          >
            {{ userInfo.name }}
          </router-link>

          <RouterLink
            v-else
            to="/auth/login"
            class="rounded-lg bg-[#e0e0e0] p-4 text-xl"
            hover="bg-[#b1aeae]"
          >
            登录
          </RouterLink>
        </ul>
      </nav>
      <nav class="group relative sm:hidden">
        <Icon width="25" height="25" icon="ion:navicon-round" />
        <div
          class="absolute z-10 hidden flex-col gap-2 rounded-lg bg-white p-4 shadow-lg -translate-x-3/5"
          group-focus="flex"
          group-hover="flex"
        >
          <router-link to="/" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            首页
          </router-link>
          

          <router-link to="/social" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            <span class="inline-flex items-center gap-2">
              <span>社交</span>
              <span v-if="totalUnread && totalUnread > 0" class="ml-2 inline-flex items-center justify-center bg-red-500 text-white text-xs font-semibold rounded-full px-2 py-0.5">{{ totalUnread > 99 ? '99+' : totalUnread }}</span>
            </span>
          </router-link>

          <!-- 修改点 2: 同样，将移动端的 <a> 标签改为 <router-link> -->
          <router-link
            v-if="userInfo?.name"
            to="/profile/info"
            class="rounded-lg bg-[#e0e0e0] p-4 text-xl"
            hover="bg-[#b1aeae]"
          >
            {{ userInfo.name }}
          </router-link>

          <RouterLink
            v-else
            to="/auth/login"
            class="rounded-lg bg-[#e0e0e0] p-4 text-xl"
            hover="bg-[#b1aeae]"
          >
            登录
          </RouterLink>
        </div>
      </nav>
    </header>
    <main class="flex-1 overflow-hidden flex flex-col">
      <RouterView />
    </main>
  </div>
</template>
