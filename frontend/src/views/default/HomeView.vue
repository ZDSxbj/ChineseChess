<script lang="ts" setup>
import type { WebSocketService } from '@/websocket'
import { inject, onMounted, onUnmounted } from 'vue'
import channel from '@/utils/channel'

const ws = inject('ws') as WebSocketService

function singlePlay() {
  channel.emit('MATCH:SUCCESS', null)
}

function onlinePlay() {
  ws?.match()
}
function handlePopState(_event: PopStateEvent) {
  window.history.pushState(null, '', window.location.href)
}
onMounted(() => {
  window.history.pushState(null, '', window.location.href)
  // 监听 popstate 事件，防止后退操作
  window.addEventListener('popstate', (event) => {
    handlePopState(event)
  })
})
onUnmounted(() => {
  window.removeEventListener('popstate', handlePopState)
})
</script>

<template>
  <main class="mx-a h-full w-9/10 bg-gray-4 p-1 sm:w-3/5">
    <h2 class="mx-a block w-fit">
      普普通通的首页
    </h2>
    <article class="text-xl line-height-9">
      <p>
        象棋与国际象棋及围棋并列世界三大棋类之一。象棋主要流行于全球华人、越南人及琉球人社区，是首届世界智力运动会的正式比赛项目之一。
      </p>
      <p>
        本网站是一个简易的象棋对战平台，支持本地对战和在线对战。在线对战需要注册账号，或者以游客登录。
      </p>
      <p>更多功能待开发...也可能不会</p>
      <p>点击左边的按钮本地对战，点击右边的按钮进行随机匹配。</p>
    </article>
    <div class="mt-4 flex flex-row">
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        @click="singlePlay"
      >
        本地对战
      </button>
      <button
        class="mx-a border-0 rounded-2xl bg-gray-2 p-4 transition-all duration-200"
        text="black xl"
        hover="bg-gray-9 text-gray-2"
        @click="onlinePlay"
      >
        随机匹配
      </button>
    </div>
  </main>
</template>
