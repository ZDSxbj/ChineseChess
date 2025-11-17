/* eslint-disable import/no-duplicates */
// ChineseChess/frontend/src/store/gameStore.ts
import type { ChessColor, ChessPiece, ChessRole } from '@/composables/ChessPiece'
import type { ChessPosition } from '@/composables/ChessPiece'

// 定义游戏状态类型（无需修改）
export interface GameState {
  isNetPlay: boolean
  roomId?: number // 联机房间ID
  selfColor: ChessColor
  moveHistory: Array<{
    from: ChessPosition
    to: ChessPosition
    capturedPiece: ChessPiece | null
    currentRole: ChessRole
  }>
  currentRole: ChessRole
}

// 保存游戏状态到sessionStorage
export function saveGameState(state: GameState) {
  try {
    // 仅将localStorage改为sessionStorage
    sessionStorage.setItem('chineseChessState', JSON.stringify(state))
  }
  catch (error) {
    console.error('Failed to save game state:', error)
  }
}

// 从sessionStorage获取游戏状态
export function getGameState(): GameState | null {
  try {
    // 仅将localStorage改为sessionStorage
    const stored = sessionStorage.getItem('chineseChessState')
    if (!stored)
      return null
    return JSON.parse(stored) as GameState
  }
  catch (error) {
    console.error('Failed to get game state:', error)
    return null
  }
}

// 清除sessionStorage中的游戏状态
export function clearGameState() {
  try {
    // 仅将localStorage改为sessionStorage
    sessionStorage.removeItem('chineseChessState')
  }
  catch (error) {
    console.error('Failed to clear game state:', error)
  }
}
