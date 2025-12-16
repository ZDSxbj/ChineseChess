<script lang="ts" setup>
import type { GameRecordItem } from '@/api/user/getGameRecords'
import type { ChessRole } from '@/store/gameStore'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getGameRecords } from '@/api/user/getGameRecords'
import { showMsg } from '@/components/MessageBox'
import { saveGameState } from '@/store/gameStore'

const router = useRouter()
const records = ref<GameRecordItem[]>([])
const loading = ref(false)

// 获取结果文本
function getResultText(result: number): string {
  switch (result) {
    case 0: return '胜'
    case 1: return '负'
    case 2: return '和'
    default: return '未知'
  }
}

// 获取结果样式类
function getResultClass(result: number): string {
  switch (result) {
    case 0: return 'result-win'
    case 1: return 'result-lose'
    case 2: return 'result-draw'
    default: return ''
  }
}

// 获取对战类型文本
function getGameTypeText(gameType: number): string {
  switch (gameType) {
    case 0: return '随机匹配'
    case 1: return '人机对战'
    case 2: return '好友对战'
    default: return '未知'
  }
}

// 获取执子颜色文本
function getColorText(isRed: boolean): string {
  return isRed ? '红方' : '黑方'
}

// 格式化时间
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 点击卡片进入复盘
function enterReplay(record: GameRecordItem) {
  try {
    // 解析 history 字段，兼容多种后端格式（JSON moves、JSON positions、紧凑数字字符串）
    const parseHistory = (raw: string) => {
      if (!raw) return []
      // 尝试解析为 JSON
      try {
        const parsed = JSON.parse(raw)
        // parsed 可能是 moves（[{from:{x,y},to:{x,y}}]) 或 positions ([{x,y},...])
        if (Array.isArray(parsed)) {
          if (parsed.length === 0) return []
          const first = parsed[0]
          // moves 格式
          if (first && typeof first === 'object' && ('from' in first) && ('to' in first)) {
            return parsed.map((m: any) => ({ from: m.from, to: m.to, capturedPiece: null, currentRole: 'self' }))
          }
          // positions 格式 -> 两个 position 为一手棋
          if (first && typeof first === 'object' && ('x' in first) && ('y' in first)) {
            const positions = parsed as Array<any>
            const moves: any[] = []
            for (let i = 0; i + 1 < positions.length; i += 2) {
              moves.push({ from: positions[i], to: positions[i + 1], capturedPiece: null, currentRole: 'self' })
            }
            return moves
          }
          // 其他数组，尝试按 pairs 解析
          const movesFallback: any[] = []
          for (let i = 0; i + 1 < parsed.length; i += 2) {
            movesFallback.push({ from: parsed[i], to: parsed[i + 1], capturedPiece: null, currentRole: 'self' })
          }
          return movesFallback
        }
      }
      catch (e) {
        // not json -> fallthrough to compact string parsing
      }

      // 紧凑字符串格式，支持两种：
      // 1. 旧格式：每4个字符表示一步 "6665" 表示 from:(6,6) to:(6,5)
      // 2. 新格式：每5个字符表示一步 "6665r" 表示红方从(6,6)走到(6,5)，最后一位是颜色标记(r/b)
      const s = raw.trim()
      const moves: any[] = []

      // 检测格式：如果第5个字符是r或b，则为新格式
      const hasColorCode = s.length >= 5 && (s[4] === 'r' || s[4] === 'b')
      const stepSize = hasColorCode ? 5 : 4

      for (let i = 0; i + 3 < s.length; i += stepSize) {
        const fromX = Number(s[i])
        const fromY = Number(s[i + 1])
        const toX = Number(s[i + 2])
        const toY = Number(s[i + 3])
        const colorCode = hasColorCode && i + 4 < s.length ? s[i + 4] : null

        if (!Number.isNaN(fromX) && !Number.isNaN(fromY) && !Number.isNaN(toX) && !Number.isNaN(toY)) {
          moves.push({
            from: { x: fromX, y: fromY },
            to: { x: toX, y: toY },
            capturedPiece: null,
            currentRole: 'self',
            pieceColor: colorCode === 'r' ? 'red' : colorCode === 'b' ? 'black' : undefined,
          })
        }
      }
      return moves
    }

    let moveHistory = parseHistory(record.history || '')
    // 注意：不需要翻转坐标！
    // 后端保存时已经根据玩家的实际视角记录了坐标
    // 复盘时以玩家自己的棋色为selfColor展示即可
    // 坐标翻转会导致复盘内容与实际对局不一致

    // 保存游戏状态到 sessionStorage，供复盘页面使用
    // 注意：GameState.currentRole 类型是 ChessRole ('self'|'enemy')，这里使用棋色计算初始显示方向（复盘以我方颜色为 SelfColor）
    saveGameState({
      isNetPlay: false,
      selfColor: record.is_red ? 'red' : 'black',
      moveHistory,
      currentRole: (record.is_red ? 'self' : 'enemy') as ChessRole,
    })

    // 跳转到复盘页面（从个人中心的记录）
    router.push({
      name: 'profile-replay',
      params: { recordId: record.id },
    })
  }
  catch (error) {
    console.error('进入复盘失败:', error)
    showMsg('进入复盘失败，请稍后重试')
  }
}

// 加载对局记录
async function loadRecords() {
  loading.value = true
  try {
    const response = await getGameRecords()
    records.value = response.records || []
    // 处理 AI 对手显示：如果是人机对战，使用后端返回的 ai_level 生成友好标签
    records.value.forEach((rec: any) => {
      if (rec.game_type === 1) {
        const lvl = rec.ai_level || 3
        let label: string
        if (lvl <= 2) {
          label = '简单'
        } else if (lvl <= 4) {
          label = '中等'
        } else {
          label = '困难'
        }
        rec.opponent_name = `AI (${label})`
      }
    })
  }
  catch (error: any) {
    console.error('加载对局记录失败:', error)
    showMsg(error.message || '加载对局记录失败')
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  loadRecords()
})
</script>

<template>
  <div class="max-w-5xl mx-auto p-8 animate-fade-in">
    <h2 class="text-2xl font-black text-amber-900 mb-8 flex items-center gap-3">
      <div class="i-carbon-game-console text-3xl"></div>
      对局记录
    </h2>

    <div v-if="loading" class="flex flex-col items-center justify-center py-20 text-amber-800/60">
      <div class="w-12 h-12 border-4 border-amber-200 border-t-amber-600 rounded-full animate-spin mb-4"></div>
      <p class="font-medium">正在加载记录...</p>
    </div>

    <div v-else-if="records.length === 0" class="text-center py-20 bg-white/40 rounded-2xl border border-amber-100">
      <div class="i-carbon-catalog text-6xl text-amber-200 mb-4 mx-auto"></div>
      <p class="text-amber-800/60 font-medium">暂无对局记录</p>
    </div>

<div v-else class="flex flex-col space-y-4">
      <div
        v-for="record in records"
        :key="record.id"
        class="group relative bg-white border border-amber-100 rounded-xl p-0 shadow-sm hover:shadow-md transition-all duration-300 cursor-pointer overflow-hidden"
        @click="enterReplay(record)"
      >
        <!-- 左侧装饰条 -->
        <div
          class="absolute left-0 top-0 bottom-0 w-1.5 transition-colors duration-300"
          :class="{
            'bg-orange-500': record.result === 0,
            'bg-gray-400': record.result === 1,
            'bg-amber-300': record.result === 2
          }"
        ></div>

        <div class="flex items-center p-4 pl-6">
          <!-- 结果与类型 -->
          <div class="flex flex-col items-center mr-6 min-w-[80px]">
            <span
              class="text-xl font-black tracking-wider"
              :class="{
                'text-orange-600': record.result === 0,
                'text-gray-500': record.result === 1,
                'text-amber-600': record.result === 2
              }"
            >
              {{ getResultText(record.result) }}
            </span>
            <span class="text-xs text-amber-800/60 mt-1 bg-amber-50 px-2 py-0.5 rounded border border-amber-100">
              {{ getGameTypeText(record.game_type) }}
            </span>
          </div>

          <!-- 分割线 -->
          <div class="w-px h-10 bg-amber-100 mr-6 hidden sm:block"></div>

          <!-- 主要信息网格 -->
          <div class="flex-1 grid grid-cols-2 sm:grid-cols-4 gap-4 items-center">
            <!-- 执子 -->
            <div class="flex flex-col">
              <span class="text-xs text-amber-800/40 mb-0.5 font-medium">执子</span>
              <span class="font-bold flex items-center gap-1.5 text-amber-900">
                <span class="w-3 h-3 rounded-full shadow-sm border border-black/10" :class="record.is_red ? 'bg-red-500' : 'bg-gray-800'"></span>
                {{ getColorText(record.is_red) }}
              </span>
            </div>

            <!-- 对手 -->
            <div class="flex flex-col">
              <span class="text-xs text-amber-800/40 mb-0.5 font-medium">对手</span>
              <span class="font-bold text-amber-900 truncate max-w-[120px]">{{ record.opponent_name }}</span>
            </div>

            <!-- 步数 -->
            <div class="flex flex-col">
              <span class="text-xs text-amber-800/40 mb-0.5 font-medium">步数</span>
              <span class="font-mono font-bold text-amber-900">{{ record.total_steps }}</span>
            </div>

            <!-- 时间 -->
            <div class="flex flex-col">
              <span class="text-xs text-amber-800/40 mb-0.5 font-medium">时间</span>
              <span class="text-xs text-amber-900/70 font-mono">{{ formatDate(record.start_time) }}</span>
            </div>
          </div>

          <!-- 箭头 -->
          <div class="ml-4 text-amber-200 group-hover:text-amber-500 transition-colors transform group-hover:translate-x-1 duration-300">
            <div class="i-carbon-chevron-right text-2xl"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


