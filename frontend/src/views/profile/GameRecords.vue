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

      // 紧凑字符串格式，例如 "6665" 表示 from:(6,6) to:(6,5)，每4个字符表示一步棋
      const s = raw.trim()
      const moves: any[] = []
      // 每4个字符为一步：fromX, fromY, toX, toY
      for (let i = 0; i + 3 < s.length; i += 4) {
        const fromX = Number(s[i])
        const fromY = Number(s[i + 1])
        const toX = Number(s[i + 2])
        const toY = Number(s[i + 3])
        if (!Number.isNaN(fromX) && !Number.isNaN(fromY) && !Number.isNaN(toX) && !Number.isNaN(toY)) {
          moves.push({
            from: { x: fromX, y: fromY },
            to: { x: toX, y: toY },
            capturedPiece: null,
            currentRole: 'self',
          })
        }
      }
      return moves
    }

    let moveHistory = parseHistory(record.history || '')
    // 如果后端保存的是统一的“红方视角”坐标，当查看者是黑方时，需要把坐标翻转到黑方视角
    if (!record.is_red && moveHistory && moveHistory.length > 0) {
      moveHistory = moveHistory.map((m: any) => ({
        from: { x: 8 - m.from.x, y: 9 - m.from.y },
        to: { x: 8 - m.to.x, y: 9 - m.to.y },
        capturedPiece: m.capturedPiece || null,
        currentRole: m.currentRole || 'self',
      }))
    }

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
    // 处理 AI 前缀：如果是人机对战且对手未知，尝试从 history 前缀读取 AI 难度并替换对手显示
    records.value.forEach((rec: any) => {
      if (rec.game_type === 1 && (!rec.opponent_name || rec.opponent_name === '未知玩家')) {
        const h: string = rec.history || ''
        const m = h.match(/^AI\|lvl=(\d+)\|/)
        if (m) {
          const lvl = Number(m[1]) || 3
          let label = '中等'
          if (lvl <= 2) label = '简单'
          else if (lvl <= 4) label = '中等'
          else label = '困难'
          rec.opponent_name = `AI (难度: ${label})`
          // 去除前缀，保持后续复盘解析兼容
          rec.history = h.replace(/^AI\|lvl=\d+\|/, '')
        } else {
          // fallback: 显示 AI
          rec.opponent_name = 'AI'
        }
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
  <div class="game-records-container">
    <h2 class="title">对局记录</h2>

    <div v-if="loading" class="loading">
      加载中...
    </div>

    <div v-else-if="records.length === 0" class="empty">
      暂无对局记录
    </div>

    <div v-else class="records-list">
      <div
        v-for="record in records"
        :key="record.id"
        class="record-card"
        @click="enterReplay(record)"
      >
        <div class="card-header">
          <div class="result-badge" :class="getResultClass(record.result)">
            {{ getResultText(record.result) }}
          </div>
          <div class="game-type">{{ getGameTypeText(record.game_type) }}</div>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="label">执子：</span>
            <span class="value" :class="record.is_red ? 'text-red' : 'text-black'">
              {{ getColorText(record.is_red) }}
            </span>
          </div>

          <div class="info-row">
            <span class="label">对手：</span>
            <span class="value">{{ record.opponent_name }}</span>
          </div>

          <div class="info-row">
            <span class="label">总步数：</span>
            <span class="value">{{ record.total_steps }}</span>
          </div>

          <div class="info-row">
            <span class="label">时间：</span>
            <span class="value time">{{ formatDate(record.start_time) }}</span>
          </div>
        </div>

        <div class="card-footer">
          <span class="replay-hint">点击查看复盘 →</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.game-records-container {
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.title {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 20px;
  color: #333;
  text-align: center;
}

.loading,
.empty {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 16px;
}

.records-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.record-card {
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.record-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
  border-color: #4a90e2;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.result-badge {
  font-size: 18px;
  font-weight: bold;
  padding: 4px 12px;
  border-radius: 4px;
  min-width: 50px;
  text-align: center;
}

.result-win {
  background: #e8f5e9;
  color: #2e7d32;
}

.result-lose {
  background: #ffebee;
  color: #c62828;
}

.result-draw {
  background: #fff3e0;
  color: #ef6c00;
}

.game-type {
  font-size: 14px;
  color: #666;
  background: #f5f5f5;
  padding: 4px 10px;
  border-radius: 4px;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.info-row {
  display: flex;
  align-items: center;
  font-size: 14px;
}

.label {
  color: #666;
  min-width: 70px;
  font-weight: 500;
}

.value {
  color: #333;
}

.text-red {
  color: #d32f2f;
  font-weight: bold;
}

.text-black {
  color: #424242;
  font-weight: bold;
}

.time {
  font-size: 13px;
  color: #999;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.replay-hint {
  font-size: 13px;
  color: #4a90e2;
  font-weight: 500;
}

.record-card:hover .replay-hint {
  color: #2a70c2;
}
</style>
