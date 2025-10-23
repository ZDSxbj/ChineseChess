<script lang="ts" setup>
import type { RoomInfo } from '@/api/room/room'
import type { WebSocketService } from '@/websocket'
import { Icon } from '@iconify/vue/dist/iconify.js'
import { inject, onMounted, ref } from 'vue'
import { getSpareRoom } from '@/api/room/get_spare_rooms'

const ws = inject('ws') as WebSocketService

const spareRoom = ref<RoomInfo[]>([])
const isLoading = ref(true)

function joinRoom(id: number) {
  ws?.join(id)
}

function createRoom() {
  ws?.create().then(() => {
    fetchRoom()
  })
}

async function fetchRoom() {
  isLoading.value = true
  try {
    const data = await getSpareRoom()
    const { rooms } = data
    spareRoom.value = rooms || []
  }
  catch (error) {
    console.error('Error fetching spare rooms:', error)
  }
  finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchRoom()
})
</script>

<template>
  <div class="mx-auto max-w-5xl w-full flex flex-col p-4">
    <h2 class="mb-6 text-center text-2xl font-bold">
      空闲房间
    </h2>

    <!-- Loading state -->
    <div v-if="isLoading" class="h-60 flex items-center justify-center">
      <div class="flex flex-col items-center gap-4">
        <div
          class="h-12 w-12 animate-spin border-4 border-gray-300 border-t-blue-500 rounded-full"
        />
        <p class="text-gray-600">
          加载中...
        </p>
      </div>
    </div>

    <!-- No rooms state -->
    <div
      v-else-if="spareRoom.length === 0"
      class="h-60 flex flex-col items-center justify-center gap-4"
    >
      <Icon icon="mdi:chess-pawn" width="48" height="48" class="text-gray-400" />
      <p class="text-center text-gray-500">
        暂无空闲房间
      </p>
    </div>

    <!-- Rooms list -->
    <div v-else class="grid grid-cols-1 gap-4 lg:grid-cols-3 md:grid-cols-2">
      <div
        v-for="room in spareRoom"
        :key="room.id"
        class="flex flex-col border border-gray-200 rounded-lg p-4 shadow-sm transition-all"
        hover="shadow-md transform scale-102"
      >
        <div class="mb-4 flex items-center justify-between">
          <h3 class="text-lg font-bold">
            房间 #{{ room.id }}
          </h3>
          <span class="rounded-full bg-green-100 px-2 py-1 text-sm text-green-800">空闲</span>
        </div>

        <div class="mb-4 flex-1">
          <!-- Current player -->
          <div class="mb-3">
            <p class="mb-1 text-sm text-gray-500">
              当前玩家
            </p>
            <div class="flex items-center gap-2">
              <div
                class="h-8 w-8 flex items-center justify-center rounded-full bg-blue-100 text-blue-700"
              >
                {{ room.current.name.charAt(0) }}
              </div>
              <div>
                <p class="font-medium">
                  {{ room.current.name }}
                </p>
                <p class="text-xs text-gray-500">
                  经验值: {{ room.current.exp }}
                </p>
              </div>
            </div>
          </div>

          <!-- Next player -->
          <div v-if="room.next && room.next.id">
            <p class="mb-1 text-sm text-gray-500">
              下一位玩家
            </p>
            <div class="flex items-center gap-2">
              <div
                class="h-8 w-8 flex items-center justify-center rounded-full bg-red-100 text-red-700"
              >
                {{ room.next.name.charAt(0) }}
              </div>
              <div>
                <p class="font-medium">
                  {{ room.next.name }}
                </p>
                <p class="text-xs text-gray-500">
                  经验值: {{ room.next.exp }}
                </p>
              </div>
            </div>
          </div>
          <div v-else class="py-2">
            <p class="text-sm text-gray-400">
              等待下一位玩家加入...
            </p>
          </div>
        </div>

        <button
          class="w-full rounded-lg bg-[#e0e0e0] py-2 text-center transition-colors"
          hover="bg-[#b1aeae]"
          @click="joinRoom(room.id)"
        >
          加入房间
        </button>
      </div>
    </div>

    <!-- Pagination or load more (optional) -->
    <div v-if="!isLoading && spareRoom.length > 0" class="mt-8 flex justify-center">
      <button
        class="border border-gray-300 rounded-lg px-4 py-2 text-gray-600 transition-colors"
        hover="bg-gray-100"
      >
        加载更多
      </button>
    </div>

    <!-- Create room button -->
    <div class="mt-8 flex justify-center">
      <button
        class="rounded-lg bg-blue-500 px-6 py-3 text-white shadow-md transition-colors"
        hover="bg-blue-600"
        @click="createRoom"
      >
        创建新房间
      </button>
    </div>
  </div>
</template>
