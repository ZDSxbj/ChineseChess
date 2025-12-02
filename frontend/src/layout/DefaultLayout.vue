<script lang="ts" setup>
import { Icon } from '@iconify/vue/dist/iconify.js'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/store/useStore'

const userStore = useUserStore()
const { userInfo } = storeToRefs(userStore)
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
          <router-link to="/about" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            关于
          </router-link>
          <router-link to="/room" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            大厅
          </router-link>
          <router-link to="/social" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            社交
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
          <router-link to="/about" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            关于
          </router-link>
          <router-link to="/room" class="rounded-lg bg-[#e0e0e0] p-4 text-xl" hover="bg-[#b1aeae]">
            大厅
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
    <RouterView />
  </div>
</template>
