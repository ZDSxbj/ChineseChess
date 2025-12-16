import type { Board, ChessColor, ChessPiece, ChessPosition, ChessRole } from './ChessPiece'
import type { GameState } from '@/store/gameStore';
import { showMsg } from '@/components/MessageBox'
import { saveGameState, getGameState, clearGameState } from '@/store/gameStore'
import channel from '@/utils/channel'
import { ChessFactory, King } from './ChessPiece'
import Drawer from './drawer'

declare const window: any

type CustomPieceType = 'King' | 'Advisor' | 'Bishop' | 'Horse' | 'Rook' | 'Cannon' | 'Pawn'
export interface CustomPieceDefinition {
  type: CustomPieceType
  color: ChessColor
  position: ChessPosition
}
export interface EndgameLayout {
  pieces: CustomPieceDefinition[]
  currentTurn: ChessColor
  selfColor?: ChessColor
  aiMode?: boolean
  skipValidation?: boolean
}

class ChessBoard {
  private board: Board
  private gridSize: number
  // 棋盘
  private boardElement: HTMLCanvasElement
  private background: CanvasRenderingContext2D
  // 棋子
  private chessesElement: HTMLCanvasElement
  private chesses: CanvasRenderingContext2D
  private selectedPiece: ChessPiece | null = null
  private selfColor: ChessColor = 'red'
  public currentRole: ChessRole = 'self'
  private isNetPlay: boolean = false
  private isAIPlay: boolean = false // 新增：标识是否为AI对战模式
  private isEndgameMode: boolean = false
  private clickCallback: (event: MouseEvent) => void = () => {}
  private eventsBound: boolean = false
  // 新增：存储走棋历史（记录移动前的状态）
  private moveHistory: Array<{
    from: ChessPosition
    to: ChessPosition
    capturedPiece: ChessPiece | null// 被吃掉的棋子（如果有）
    currentRole: ChessRole // 记录当前回合角色，用于悔棋后恢复
    pieceName?: string
    pieceColor?: ChessColor
  }> = []
  // 提示位置
  private hintPositions: ChessPosition[] = []
  // 对战音效（选中与落子）
  private selectAudio?: HTMLAudioElement
  private clickAudio?: HTMLAudioElement
  private eatAudio?: HTMLAudioElement
  private checkAudio?: HTMLAudioElement

  constructor(
    boardElement: HTMLCanvasElement,
    chessesElement: HTMLCanvasElement,
    gridSize: number = 50,
  ) {
    this.boardElement = boardElement
    this.chessesElement = chessesElement
    this.gridSize = gridSize
    this.board = Array.from({ length: 9 }).fill(null).map(() => {
      return {}
    })
    this.initCanvasElement()
    this.background = this.boardElement.getContext('2d') as CanvasRenderingContext2D
    this.chesses = this.chessesElement.getContext('2d') as CanvasRenderingContext2D
    // 音频初始化（资源路径位于 public/audio/ 下）
    try {
      this.selectAudio = new Audio('/chess/audio/select.wav')
      this.selectAudio.preload = 'auto'
      this.selectAudio.volume = 0.7
      this.clickAudio = new Audio('/chess/audio/click.wav')
      this.clickAudio.preload = 'auto'
      this.clickAudio.volume = 0.7
      // 新增：吃子与将军音效
      this.eatAudio = new Audio('/chess/audio/eat.mp3')
      this.eatAudio.preload = 'auto'
      this.eatAudio.volume = 0.8
      this.checkAudio = new Audio('/chess/audio/check.mp3')
      this.checkAudio.preload = 'auto'
      this.checkAudio.volume = 0.8
    } catch (e) {
      console.warn('Audio init failed', e)
    }
    this.listenEvent()
  }

  // Public API: 在本地模式下应用来自 AI 的移动（from/to 使用内部坐标 {x,y}）
  public applyAIMove(from: ChessPosition, to: ChessPosition): boolean {
    console.log('applyAIMove called with', from, to)
    const piece = this.board[from.x] && this.board[from.x][from.y]
    console.log('piece at from coordinate:', piece)
    const targetPiece = this.board[to.x][to.y]
    if (!piece) {
      console.error('applyAIMove failed: no piece at from position', from)
      return false
    }

    // 防止自吃
    if (targetPiece && targetPiece.color === piece.color) {
      console.error('applyAIMove failed: target occupied by same color', { from, to, piece, targetPiece })
      return false
    }

    if (!piece.isMoveValid(to, this.board)) {
      console.error('applyAIMove failed: piece.isMoveValid returned false', { pieceName: piece.name, from: piece.position, to })
      return false
    }

    // 模拟移动检查自将
    const savedPiece = this.board[to.x][to.y]
    const originalPos = { ...piece.position }
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    piece.position = to

    const wouldBeInCheck = this.isInCheck(piece.color)

    // 恢复
    piece.position = originalPos
    this.board[from.x][from.y] = piece
    if (savedPiece) this.board[to.x][to.y] = savedPiece
    else delete this.board[to.x][to.y]

    if (wouldBeInCheck) {
      console.error('applyAIMove failed: move would leave mover in check', { piece: piece.name, from, to })
      return false
    }

    // 记录历史并应用移动（包含棋子名与阵营，便于 UI 恢复显示）
    this.moveHistory.push({ from: { ...from }, to: { ...to }, capturedPiece: targetPiece || null, currentRole: this.currentRole, pieceName: piece.name, pieceColor: piece.color })
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    piece.move(to)
    this.drawChesses()

    // 发出最近一步走子事件（包含棋子名和阵营），以便 UI 显示最近落子
    try {
      channel.emit('BOARD:MOVE:MADE', { from, to, pieceName: piece.name, pieceColor: piece.color })
    } catch (e) {
      console.warn('BOARD:MOVE:MADE emit failed', e)
    }

    // 检查对方被将军或将死
    const opponentColor = piece.color === 'red' ? 'black' : 'red'
    const opponentInCheck = this.isInCheck(opponentColor)
    // 本地规则判断
    const opponentCheckmatedLocal = opponentInCheck && this.isCheckmate(opponentColor)
    // 使用 engine 作为补充判断（若存在），解决规则判定差异问题
    let opponentCheckmatedEngine = false
    try {
      // @ts-ignore
      if ((window as any).logic && typeof (window as any).logic.isCheckmate === 'function') {
        // engine 接受 board[row][col] 的形式
        opponentCheckmatedEngine = (window as any).logic.isCheckmate(this.getCurrentBoard(), opponentColor)
      }
    } catch (e) {
      console.warn('Engine checkmate validation failed', e)
    }
    const opponentCheckmated = opponentCheckmatedLocal || opponentCheckmatedEngine

    if (opponentCheckmated) {
      const moverColor = piece.color
      if (this.isNetPlay) channel.emit('GAME:END', { winner: moverColor, online: true })
      else channel.emit('LOCAL:GAME:END', { winner: moverColor })
      this.end(moverColor)
      return true
    } else if (opponentInCheck) {
        if (this.isAIPlay) {
          if (opponentColor === this.selfColor) {
            showMsg('你被将军了！')
          } else {
            showMsg('AI被将军！')
          }
        } else {
          showMsg(`${opponentColor === 'red' ? '红方' : '黑方'}被将军！`)
        }
    }

    // 切换回合（AI/本地/网络统一行为）
    this.currentRole = this.currentRole === 'self' ? 'enemy' : 'self'
    // 通知 UI 当前回合改变
    try {
      channel.emit('BOARD:ROLE:CHANGE', { currentRole: this.currentRole })
    } catch (e) {
      console.warn('BOARD:ROLE:CHANGE emit failed', e)
    }
    saveGameState({ isNetPlay: this.isNetPlay, selfColor: this.selfColor, moveHistory: this.moveHistory, currentRole: this.currentRole })
    return true
  }

  // 返回指定坐标的棋子（若无返回 null）
  public getPieceAt(pos: ChessPosition) {
    return (this.board[pos.x] && this.board[pos.x][pos.y]) || null
  }

  public isNetworkPlay(): boolean {
    return this.isNetPlay
  }

  get stepsNum(): number {
    return this.moveHistory.length
  }

  // 检测指定颜色的将/帅是否被将军
  private isInCheck(color: ChessColor): boolean {
    // 找到该颜色的将/帅
    let king: { piece: ChessPiece, pos: ChessPosition } | null = null
    for (let x = 0; x <= 8; x++) {
      for (let y = 0; y <= 9; y++) {
        const piece = this.board[x][y]
        if (piece && piece instanceof King && piece.color === color) {
          king = { piece, pos: { x, y } }
          break
        }
      }
      if (king) break
    }
    if (!king) return false

    // 首先检查将帅是否碰面（同列且中间无棋子）
    const enemyColor = color === 'red' ? 'black' : 'red'
    let enemyKing: { piece: ChessPiece, pos: ChessPosition } | null = null
    for (let x = 0; x <= 8; x++) {
      for (let y = 0; y <= 9; y++) {
        const piece = this.board[x][y]
        if (piece && piece instanceof King && piece.color === enemyColor) {
          enemyKing = { piece, pos: { x, y } }
          break
        }
      }
      if (enemyKing) break
    }

    // 如果找到对方的将/帅，检查是否碰面
    if (enemyKing && king.pos.x === enemyKing.pos.x) {
      let blocked = false
      for (let yi = Math.min(king.pos.y, enemyKing.pos.y) + 1; yi < Math.max(king.pos.y, enemyKing.pos.y); yi++) {
        if (this.board[king.pos.x][yi]) {
          blocked = true
          break
        }
      }
      // 如果将帅在同一列且中间无棋子，则被将军
      if (!blocked) {
        return true
      }
    }

    // 检查所有对方棋子是否能攻击到这个将/帅
    for (let x = 0; x <= 8; x++) {
      for (let y = 0; y <= 9; y++) {
        const piece = this.board[x][y]
        if (piece && piece.color === enemyColor) {
          // 检查这个敌方棋子能否移动到将/帅的位置
          if (piece.isMoveValid(king.pos, this.board)) {
            return true
          }
        }
      }
    }
    return false
  }

  // 检测指定颜色是否可以通过任何合法移动解除将军
  private canEscapeCheck(color: ChessColor): boolean {
    // 遍历所有己方棋子，尝试所有可能的移动
    for (let fromX = 0; fromX <= 8; fromX++) {
      for (let fromY = 0; fromY <= 9; fromY++) {
        const piece = this.board[fromX][fromY]
        if (!piece || piece.color !== color) continue

        // 尝试所有可能的目标位置
        for (let toX = 0; toX <= 8; toX++) {
          for (let toY = 0; toY <= 9; toY++) {
            const targetPiece = this.board[toX][toY]
            // 不能吃自己的子
            if (targetPiece && targetPiece.color === color) continue

            // 检查这步棋是否合法
            if (!piece.isMoveValid({ x: toX, y: toY }, this.board)) continue

            // 模拟这步棋
            const savedPiece = this.board[toX][toY]
            delete this.board[fromX][fromY]
            this.board[toX][toY] = piece
            const savedPos = { ...piece.position }
            piece.position = { x: toX, y: toY }

            // 检查移动后是否还在将军状态
            const stillInCheck = this.isInCheck(color)

            // 恢复棋盘状态
            piece.position = savedPos
            this.board[fromX][fromY] = piece
            if (savedPiece) {
              this.board[toX][toY] = savedPiece
            } else {
              delete this.board[toX][toY]
            }

            // 如果这步棋能解除将军，返回true
            if (!stillInCheck) {
              return true
            }
          }
        }
      }
    }
    return false
  }

  // 检测指定颜色是否被将死
  public isCheckmate(color: ChessColor): boolean {
    return this.isInCheck(color) && !this.canEscapeCheck(color)
  }

  // 检查当前回合方是否被将死，如果是则结束游戏
  private checkCurrentTurnCheckmate() {
    // 联机模式下由服务端/对手走子触发判定，本地模式才需要主动检测
    if (this.isNetPlay) return false
    // 确定当前回合应该走棋的颜色
    const enemyColor = this.selfColor === 'red' ? 'black' : 'red'
    const currentColor = this.currentRole === 'self' ? this.selfColor : enemyColor
    
    // 先用本地判定
    let currentIsCheckmatedLocal = this.isCheckmate(currentColor)
    let currentIsCheckmatedEngine = false
    try {
      // @ts-ignore
      if ((window as any).logic && typeof (window as any).logic.isCheckmate === 'function') {
        currentIsCheckmatedEngine = (window as any).logic.isCheckmate(this.getCurrentBoard(), currentColor)
      }
    } catch (e) {
      console.warn('Engine checkmate validation failed', e)
    }

    if (currentIsCheckmatedLocal || currentIsCheckmatedEngine) {
      // 当前回合方被将死，对方获胜
      const winner = currentColor === 'red' ? 'black' : 'red'
      showMsg(`${currentColor === 'red' ? '红方' : '黑方'}被将死！`)
      // 本地对战：仅通知本地 UI
      channel.emit('LOCAL:GAME:END', { winner })
      this.end(winner)
      return true
    }
    return false
  }

  get width(): number {
    return this.gridSize * 9
  }

  get height(): number {
    return this.gridSize * 10
  }

  get Color(): ChessColor {
    return this.selfColor
  }

  get moveHistoryList() {
    return this.moveHistory
  }

  get SelfColor(): ChessColor {
    return this.selfColor
  }

  private clickHandler(event: MouseEvent) {
    // 在允许任何操作前，检查当前回合方是否已被将死
    if (this.checkCurrentTurnCheckmate()) {
      return
    }

    const rect = this.chessesElement.getBoundingClientRect()
    const x = Math.floor((event.clientX - rect.left) / this.gridSize)
    const y = Math.floor((event.clientY - rect.top) / this.gridSize)

    // 自己是什么颜色
    const selfColor = this.selfColor
    // 对方是什么颜色
    const enemyColor = this.selfColor === 'red' ? 'black' : 'red'

    // 棋子点击事件
    const piece = this.board[x][y]
    if (this.selectedPiece) {
      if (piece === this.selectedPiece) {
        this.selectedPiece.deselect()
        this.selectedPiece = null
        this.hintPositions = []
        this.drawChesses()
        return
      }
      if (!piece || piece.color !== this.selectedPiece.color) {
        const curPiece = this.selectedPiece
        this.move(curPiece.position, { x, y })
        this.selectedPiece.deselect()
        this.selectedPiece = null
        this.hintPositions = []
        this.drawChesses()
        return
      }
    }

    if (piece) {
      // 不是当前角色的棋子不能选中
      if (piece.color !== (this.currentRole === 'self' ? selfColor : enemyColor)) {
        return
      }
      // 网络模式或AI模式下，不是自己的回合不能选中
      if (this.isNetPlay || this.isAIPlay) {
        if (this.currentRole === 'enemy') {
          return
        }
      }
      this.selectPiece(piece)
    }
  }

  private move(from: ChessPosition, to: ChessPosition) {
    const piece = this.board[from.x][from.y]
    const targetPiece = this.board[to.x][to.y]
    if (!piece) {
      return
    }

    // 防止自吃（特别是联机从网络通道直接触发时）
    if (targetPiece && targetPiece.color === piece.color) {
      return
    }

    if (!piece.isMoveValid(to, this.board)) {
      return
    }
    // 先在数据结构上应用移动（模拟新的棋盘状态），再绘制

    // 模拟移动，检查是否会导致自己被将军
    const savedPiece = this.board[to.x][to.y]
    const originalPos = { ...piece.position }
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    // 仅修改内存位置用于判定，不进行绘制
    piece.position = to

    const wouldBeInCheck = this.isInCheck(piece.color)

    // 恢复棋盘状态
    piece.position = originalPos
    this.board[from.x][from.y] = piece
    if (savedPiece) {
      this.board[to.x][to.y] = savedPiece
    } else {
      delete this.board[to.x][to.y]
    }

    // 如果移动后自己会被将军，则不允许这步棋
    if (wouldBeInCheck) {
      showMsg('不能送将！')
      return
    }
    // 记录历史（包含棋子名/阵营，便于落子文本恢复）
    this.moveHistory.push({
      from: { ...from },
      to: { ...to },
      capturedPiece: targetPiece || null,
      currentRole: this.currentRole, // 记录当前角色，悔棋后需恢复
      pieceName: piece.name,
      pieceColor: piece.color,
    })

    // 先在数据结构上应用移动（模拟新的棋盘状态），再绘制
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    piece.move(to)
    this.drawChesses()
    // 非吃子落子音效（仅联机时启用，先不播放吃子，等判定是否将军后再决定）
    if (this.isNetPlay && !targetPiece) {
      try {
        if (this.clickAudio) {
          this.clickAudio.currentTime = 0
          void this.clickAudio.play()
        }
      } catch {}
    }

    // 只有自己走才发送走子事件
    if (this.currentRole === 'self') {
      this.isNetPlay && channel.emit('NET:CHESS:MOVE:END', { from, to })
    }
    // 通知 UI 最近一步走子（本地/网络/AI 都需要）
    try {
      channel.emit('BOARD:MOVE:MADE', { from, to, pieceName: piece.name, pieceColor: piece.color })
    } catch (e) {
      console.warn('BOARD:MOVE:MADE emit failed', e)
    }

    // 检查对方是否被将军或将死
    const opponentColor = piece.color === 'red' ? 'black' : 'red'
    const opponentInCheck = this.isInCheck(opponentColor)
    // 本地规则判断
    const opponentCheckmatedLocal = opponentInCheck && this.isCheckmate(opponentColor)
    // 使用 engine 作为补充判断（若存在），解决规则判定差异问题
    let opponentCheckmatedEngine = false
    try {
      // @ts-ignore
      if ((window as any).logic && typeof (window as any).logic.isCheckmate === 'function') {
        opponentCheckmatedEngine = (window as any).logic.isCheckmate(this.getCurrentBoard(), opponentColor)
      }
    } catch (e) {
      console.warn('Engine checkmate validation failed', e)
    }
    const opponentCheckmated = opponentCheckmatedLocal || opponentCheckmatedEngine

    if (opponentCheckmated) {
      // 对方被将死，移动方获胜
      const moverColor = piece.color
      if (this.isNetPlay) {
        channel.emit('GAME:END', { winner: moverColor, online: true })
      } else {
        channel.emit('LOCAL:GAME:END', { winner: moverColor })
      }
      this.end(moverColor)
      return
    } else if (opponentInCheck) {
      // 对方被将军或我方被将军（取决于当前移动方）
      if (this.isAIPlay) {
        if (opponentColor === this.selfColor) {
          showMsg('你被将军了！')
        } else {
          showMsg('AI被将军！')
        }
      } else {
        showMsg(`${opponentColor === 'red' ? '红方' : '黑方'}被将军！`)
      }
      try {
        if (this.checkAudio) {
          this.checkAudio.currentTime = 0
          void this.checkAudio.play()
        }
      } catch {}
    } else {
      // 非将军的情况下：若发生吃子则播放吃子音效（仅联机时启用），否则不做将军提示
      if (this.isNetPlay && !!targetPiece) {
        try {
          if (this.eatAudio) {
            this.eatAudio.currentTime = 0
            void this.eatAudio.play()
          }
        } catch {}
      }
      // 非将军，不显示将军提示
    }

    // 如果直接吃掉对方的将，立即判定结束（移动方胜）
    if (targetPiece && targetPiece instanceof King) {
      const moverColor = piece.color
      if (this.isNetPlay) {
        channel.emit('GAME:END', { winner: moverColor, online: true })
      } else {
        channel.emit('LOCAL:GAME:END', { winner: moverColor })
      }
      this.end(moverColor)
      return
    }

    // 切换回合
    this.currentRole = this.currentRole === 'self' ? 'enemy' : 'self'
    // showMsg(`现在是${this.currentRole}的回合`)
    // 通知 UI 当前回合已改变
    try {
      channel.emit('BOARD:ROLE:CHANGE', { currentRole: this.currentRole })
    } catch (e) {
      console.warn('BOARD:ROLE:CHANGE emit failed', e)
    }
    saveGameState({
      isNetPlay: this.isNetPlay,
      selfColor: this.selfColor,
      moveHistory: this.moveHistory,
      currentRole: this.currentRole,
    })

    // 在本地模式下，切换回合后立即检查新回合方是否被将死
    // 使用 setTimeout 确保棋盘 UI 更新后再检查
    if (!this.isNetPlay) {
      setTimeout(() => {
        this.checkCurrentTurnCheckmate()
      }, 10)
      // 触发GAME:MOVE事件用于AI对战
      channel.emit('GAME:MOVE')
    }
  }

  // 实际执行悔棋的方法
  public regretMove(): boolean {
    const lastMove = this.moveHistory.pop()
    if (!lastMove)
      return false
    const { from, to, capturedPiece } = lastMove
    const piece = this.board[to.x][to.y]
    // 将棋子移回原来的位置
    piece.move(from)
    this.board[from.x][from.y] = piece
    delete this.board[to.x][to.y]
    // 恢复被吃掉的棋子
    if (capturedPiece) {
      this.board[to.x][to.y] = capturedPiece
      capturedPiece.move(to)
    }
    this.drawChesses()
    this.currentRole = this.currentRole === 'self' ? 'enemy' : 'self'
    // 通知 UI 当前回合已改变（悔棋后）
    try {
      channel.emit('BOARD:ROLE:CHANGE', { currentRole: this.currentRole })
    } catch (e) {
      console.warn('BOARD:ROLE:CHANGE emit failed', e)
    }

    saveGameState({
      isNetPlay: this.isNetPlay,
      selfColor: this.selfColor,
      moveHistory: this.moveHistory,
      currentRole: this.currentRole,
    })
    return true
  }

  // 回退到玩家上一步之前：优先撤销 AI 步与玩家步，保证回到玩家移动前的局面
  // 返回撤销的步数（0/1/2）
  public regretLastTurn(): number {
    if (!this.moveHistory || this.moveHistory.length === 0) return 0

    const last = this.moveHistory[this.moveHistory.length - 1]
    let undone = 0

    // 如果最后一步是 AI 的（enemy），尝试撤销 AI + 玩家 两步
    if (last.currentRole === 'enemy') {
      // 撤销 AI 步
      if (this.regretMove()) {
        undone++
      }
      // 再撤销玩家步（若存在）
      if (this.moveHistory.length > 0) {
        if (this.regretMove()) {
          undone++
        }
      }
    } else {
      // 最后一步是玩家的，撤销一步
      if (this.regretMove()) {
        undone++
      }
    }

    // 确保回到玩家回合
    this.currentRole = 'self'
    saveGameState({ isNetPlay: this.isNetPlay, selfColor: this.selfColor, moveHistory: this.moveHistory, currentRole: this.currentRole })
    return undone
  }

  // 新增：恢复游戏状态的方法
  public restoreState(savedState: GameState): void {
    // 恢复基础配置（覆盖初始设置）
    this.start(savedState.selfColor, savedState.isNetPlay)

    // 恢复历史记录和当前回合
    this.moveHistory = savedState.moveHistory // 注意：这里直接赋值私有属性，或通过setter
    this.currentRole = savedState.currentRole

    // 重新应用历史记录到棋盘
    this.clear() // 清空当前棋盘
    this.initChesses() // 重新初始化棋子

    // 遍历历史步骤，还原每一步移动
    savedState.moveHistory.forEach((step) => {
      const piece = this.board[step.from.x][step.from.y]
      if (piece) {
        piece.move(step.to)
        delete this.board[step.from.x][step.from.y]
        this.board[step.to.x][step.to.y] = piece
      }
    })

    this.drawChesses()
    this.drawBoard()
  }

  // 判断是否是自己的回合
  public isMyTurn(): boolean {
    return this.currentRole === 'self'
  }

  // 新增：手动设置当前执行方（用于悔棋后切换）
  public setCurrentRole(role: ChessRole) {
    this.currentRole = role
    try {
      channel.emit('BOARD:ROLE:CHANGE', { currentRole: this.currentRole })
    } catch (e) {
      console.warn('BOARD:ROLE:CHANGE emit failed', e)
    }
  }

  // 获取当前棋盘状态（用于AI计算）
  public getCurrentBoard(): any {
    const boardArray = Array.from({ length: 10 }, () => Array(9).fill(null))
    
    // 根据棋子名称映射到logic.js的简称（兼容英文与中文名称）
    const nameToAbbr: Record<string, string> = {
      // 中文
      '將': 'k', '帅': 'k', '帥': 'k',
      '士': 'a', '仕': 'a',
      '象': 'b', '相': 'b',
      '馬': 'n', '傌': 'n', '马': 'n',
      '車': 'r', '俥': 'r', '车': 'r',
      '炮': 'c', '砲': 'c',
      '卒': 's', '兵': 's',
      // 英文类名
      'King': 'k',
      'Advisor': 'a',
      'Bishop': 'b',
      'Horse': 'n',
      'Rook': 'r',
      'Cannon': 'c',
      'Pawn': 's',
    }
    
    for (let x = 0; x <= 8; x++) {
      for (let y = 0; y <= 9; y++) {
        const piece = this.board[x][y]
        if (piece) {
          // 转换为logic.js格式
          const abbr = nameToAbbr[piece.name] || 'k'
          // 如果玩家执黑（selfColor === 'black'），前端棋盘是翻转的，需要在发送到 engine 时进行纵向翻转
          const engineY = this.selfColor === 'black' ? 9 - y : y
          boardArray[engineY][x] = {
            t: abbr,
            c: piece.color,
          }
        }
      }
    }
    
    return boardArray
  }

  // 直接设置棋盘状态（用于AI走棋后更新棋盘）
  public setCurrentBoard(boardArray: any): void {
    // 清空棋盘
    for (let x = 0; x <= 8; x++) {
      for (let y = 0; y <= 9; y++) {
        delete this.board[x][y]
      }
    }

    // 根据数组重建棋盘
    for (let y = 0; y < 10; y++) {
      for (let x = 0; x < 9; x++) {
        // 如果玩家执黑，engine 的 row 需要翻回前端坐标
        const enginePiece = boardArray[y][x]
        const piece = enginePiece
        if (piece) {
          // 将 logic.js 的简称映射为 ChessFactory 可识别的英文类型名
          const abbrToName: Record<string, string> = {
            k: 'King',
            a: 'Advisor',
            b: 'Bishop',
            n: 'Horse',
            r: 'Rook',
            c: 'Cannon',
            s: 'Pawn',
          }
          const type = abbrToName[(piece.t || '').toLowerCase()] || 'King'
          const color = piece.c as any
          const role: any = color === this.selfColor ? 'self' : 'enemy'
          // 计算前端坐标（如果玩家执黑，需要翻转 engine 的行）
          const frontendY = this.selfColor === 'black' ? 9 - y : y
          // 使用 ChessFactory.createChessPiece 创建棋子实例
          const id = x * 10 + frontendY // 简单生成唯一 id
          try {
            const chessPiece = ChessFactory.createChessPiece(this.chesses, id, type as any, color, role, x, this.gridSize)
            if (chessPiece) {
              // override default constructor position so piece is placed at intended coordinates
              chessPiece.position = { x, y: frontendY }
              this.board[x][frontendY] = chessPiece
            }
          } catch (err) {
            console.error('Failed to create chess piece from boardArray:', piece, err)
          }
        }
      }
    }
  }

  // 重新渲染棋盘（用于AI走棋后刷新显示）
  public render(): void {
    this.background.clearRect(0, 0, this.width, this.height)
    this.chesses.clearRect(0, 0, this.width, this.height)

    // 先绘制棋盘，再通过统一的绘制逻辑绘制棋子/提示/最近一步高亮
    this.drawBoard()
    this.drawChesses()
  }

  private listenClick() {
    // 如果已经存在注册的回调，先移除，避免重复绑定导致后续无法正确移除
    if (this.clickCallback) {
      this.chessesElement.removeEventListener('click', this.clickCallback)
    }
    this.clickCallback = this.clickHandler.bind(this)
    this.chessesElement.addEventListener('click', this.clickCallback)
  }

  private end(_winner: string) {
    this.chessesElement.removeEventListener('click', this.clickCallback)
  }

  // 公有方法：禁用用户交互（用于网络结束时保留棋盘但禁止继续操作）
  public disableInteraction() {
    this.chessesElement.removeEventListener('click', this.clickCallback)
  }

  private isMoveSafe(piece: ChessPiece, to: ChessPosition): boolean {
    const from = piece.position
    const savedPiece = this.board[to.x][to.y]
    const originalPos = { ...piece.position }
    
    // Temporarily move
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    piece.position = to

    const wouldBeInCheck = this.isInCheck(piece.color)

    // Restore
    piece.position = originalPos
    this.board[from.x][from.y] = piece
    if (savedPiece) {
      this.board[to.x][to.y] = savedPiece
    } else {
      delete this.board[to.x][to.y]
    }

    return !wouldBeInCheck
  }

  private calculateHints(piece: ChessPiece) {
    this.hintPositions = []
    for (let x = 0; x <= 8; x++) {
      for (let y = 0; y <= 9; y++) {
        const targetPos = { x, y }
        // Check if move is valid according to piece rules
        if (!piece.isMoveValid(targetPos, this.board)) {
          continue
        }
        
        // Check if target is occupied by own piece
        const targetPiece = this.board[x][y]
        if (targetPiece && targetPiece.color === piece.color) {
          continue
        }

        // Check if move causes self-check
        if (!this.isMoveSafe(piece, targetPos)) {
          continue
        }

        this.hintPositions.push(targetPos)
      }
    }
  }

  private drawHints() {
    this.hintPositions.forEach(pos => {
      // Only draw if empty (as per user request: "if it can eat a certain piece, that piece does not need to be marked")
      if (!this.board[pos.x][pos.y]) {
        const x = pos.x * this.gridSize + this.gridSize / 2
        const y = pos.y * this.gridSize + this.gridSize / 2
        this.chesses.beginPath()
        this.chesses.arc(x, y, 5, 0, Math.PI * 2) // Small red dot
        this.chesses.fillStyle = 'red'
        this.chesses.fill()
      }
    })
  }

  private selectPiece(piece: ChessPiece) {
    if (!this.isNetPlay) {
      // 本地
      if (!piece || piece.role !== this.currentRole) {
        return
      }
    }
    else {
      // 联网
      if (piece.color !== this.selfColor) {
        return
      }
    }

    if (this.selectedPiece === piece) {
      this.selectedPiece.deselect()
      this.selectedPiece = null
      this.hintPositions = []
      this.drawChesses()
      return
    }

    this.selectedPiece?.deselect()
    this.selectedPiece = piece
    piece.select()
    
    this.calculateHints(piece)
    this.drawChesses()

    // 选中棋子音效（仅联机时启用）
    if (this.isNetPlay) {
      this.selectAudio?.play().catch(() => {})
    }
  }

  public start(color: ChessColor, isNet: boolean, isAI: boolean = false) {
    this.isNetPlay = isNet
    this.isAIPlay = isAI
    this.isEndgameMode = false
    this.selfColor = color
    this.currentRole = color === 'red' ? 'self' : 'enemy'
    this.moveHistory = []
    this.board = Array.from({ length: 9 }).fill(null).map(() => {
      return {}
    })
    this.drawBoard()
    this.initChesses()
    this.drawChesses()
    this.listenClick()
    this.listenEvent()
  }

  // Endgame entry: load custom layout with validation
  public startEndgame(layout: EndgameLayout) {
    const { pieces, currentTurn, selfColor = 'red', aiMode = true, skipValidation } = layout
    const validation = skipValidation ? { valid: true, errors: [] } : ChessBoard.validateCustomLayout(pieces)
    if (!validation.valid) {
      const reason = validation.errors.join('; ')
      console.error('[ChessBoard] Endgame layout invalid:', reason)
      showMsg(`残局布局不合法：${reason}`)
      return { success: false, reason }
    }

    this.isNetPlay = false
    this.isAIPlay = aiMode
    this.isEndgameMode = true
    this.selfColor = selfColor
    this.currentRole = currentTurn === selfColor ? 'self' : 'enemy'
    this.moveHistory = []

    // 重置棋盘画布与数据
    this.board = Array.from({ length: 9 }).map(() => ({})) as Board
    this.clear()
    this.drawBoard()

    // 将自定义棋子放置到棋盘
    this.rebuildBoardFromPieces(pieces)
    this.drawChesses()
    this.listenClick()
    this.listenEvent()
    saveGameState({ isNetPlay: false, selfColor: this.selfColor, moveHistory: [], currentRole: this.currentRole })
    return { success: true }
  }

  // 校验并快速评估残局布局是否可行
  public static assessEndgameLayout(pieces: CustomPieceDefinition[], currentTurn: ChessColor) {
    const validation = ChessBoard.validateCustomLayout(pieces)
    if (!validation.valid) return { playable: false, issues: validation.errors }
    const issues: string[] = []
    try {
      // 若引擎可用，确保先手非被将死/无着法
      const board = ChessBoard.toEngineBoardFromPieces(pieces)
      const logic = (window as any)?.logic
      if (logic && typeof logic.isCheckmate === 'function') {
        if (logic.isCheckmate(board, currentTurn)) issues.push('先手已被将死')
      }
      if (logic && typeof logic.hasAnyLegalMoves === 'function') {
        if (!logic.hasAnyLegalMoves(board, currentTurn)) issues.push('先手无合法着法')
      }
    } catch (e: any) {
      issues.push(e?.message || '布局可行性检测失败')
    }
    return { playable: issues.length === 0, issues }
  }

  // 专门用于复盘时执行的移动（只改变棋盘状态和绘制，不触发网络事件或胜负判定、回合切换等）
  public replayMove(from: ChessPosition, to: ChessPosition) {
    const piece = this.board[from.x][from.y]
    if (!piece) {
      return
    }
    // 处理被吃棋子（复盘时我们只是覆盖位置即可）
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    piece.move(to)
    // 只负责绘制，不修改历史、回合或触发事件
    this.drawChesses()
  }

  public stop() {
    this.chessesElement.removeEventListener('click', this.clickCallback)
    this.clear()
    channel.off('NET:CHESS:MOVE')
    this.eventsBound = false
  }

  private listenMove(req: { from: ChessPosition, to: ChessPosition }) {
    const { from, to } = req
    this.move(from, to)
    this.drawChesses()
  }

  private listenEvent() {
    if (this.eventsBound) return
    channel.on('NET:CHESS:MOVE', this.listenMove.bind(this))
    this.eventsBound = true
  }

  private initChesses() {
    let id = 0
    // 車 馬 象 士 炮 兵 將
    for (let i = 0; i < 2; ++i) {
      const x = i * 8 + 0
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Rook',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][9] = piece
    }

    for (let i = 0; i < 2; ++i) {
      const x = i * 6 + 1
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Horse',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][9] = piece
    }

    for (let i = 0; i < 2; ++i) {
      const x = i * 4 + 2
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Bishop',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][9] = piece
    }

    for (let i = 0; i < 2; ++i) {
      const x = i * 2 + 3
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Advisor',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][9] = piece
    }

    for (let i = 0; i < 2; ++i) {
      const x = i * 6 + 1
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Cannon',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][7] = piece
    }

    for (let i = 0; i < 5; ++i) {
      const x = i * 2 + 0
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Pawn',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][6] = piece
    }

    for (let i = 0; i < 1; ++i) {
      const x = 4
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'King',
        this.selfColor,
        'self',
        x,
        this.gridSize,
      )
      id++
      this.board[x][9] = piece
    }

    // 反方棋子
    const enemyColor = this.selfColor === 'red' ? 'black' : 'red'
    for (let i = 0; i < 2; ++i) {
      const x = i * 8 + 0
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Rook',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][0] = piece
    }
    for (let i = 0; i < 2; ++i) {
      const x = i * 6 + 1
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Horse',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][0] = piece
    }
    for (let i = 0; i < 2; ++i) {
      const x = i * 4 + 2
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Bishop',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][0] = piece
    }
    for (let i = 0; i < 2; ++i) {
      const x = i * 2 + 3
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Advisor',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][0] = piece
    }
    for (let i = 0; i < 2; ++i) {
      const x = i * 6 + 1
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Cannon',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][2] = piece
    }
    for (let i = 0; i < 5; ++i) {
      const x = i * 2 + 0
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'Pawn',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][3] = piece
    }
    for (let i = 0; i < 1; ++i) {
      const x = 4
      const piece = ChessFactory.createChessPiece(
        this.chesses,
        id,
        'King',
        enemyColor,
        'enemy',
        x,
        this.gridSize,
      )
      id++
      this.board[x][0] = piece
    }
  }

  private initCanvasElement() {
    this.boardElement.width = this.width
    this.boardElement.height = this.height
    this.chessesElement.width = this.width
    this.chessesElement.height = this.height
  }

  public redraw(newSize: number = this.gridSize) {
    this.clear()
    this.gridSize = newSize
    this.initCanvasElement()
    this.drawBoard()
    this.drawChesses()
  }

  private clear() {
    this.background.clearRect(0, 0, this.width, this.height)
    this.chesses.clearRect(0, 0, this.width, this.height)
  }

  private drawSelectionBox(x: number, y: number) {
    const ctx = this.chesses
    const gridSize = this.gridSize
    const half = gridSize / 2
    const len = gridSize / 6
    const padding = 1
    
    const centerX = x * gridSize + half
    const centerY = y * gridSize + half

    ctx.lineWidth = 3
    ctx.strokeStyle = 'red'
    ctx.lineCap = 'square'

    // Top Left
    ctx.beginPath()
    ctx.moveTo(centerX - half + padding, centerY - half + padding + len)
    ctx.lineTo(centerX - half + padding, centerY - half + padding)
    ctx.lineTo(centerX - half + padding + len, centerY - half + padding)
    ctx.stroke()

    // Top Right
    ctx.beginPath()
    ctx.moveTo(centerX + half - padding - len, centerY - half + padding)
    ctx.lineTo(centerX + half - padding, centerY - half + padding)
    ctx.lineTo(centerX + half - padding, centerY - half + padding + len)
    ctx.stroke()

    // Bottom Right
    ctx.beginPath()
    ctx.moveTo(centerX + half - padding, centerY + half - padding - len)
    ctx.lineTo(centerX + half - padding, centerY + half - padding)
    ctx.lineTo(centerX + half - padding - len, centerY + half - padding)
    ctx.stroke()

    // Bottom Left
    ctx.beginPath()
    ctx.moveTo(centerX - half + padding + len, centerY + half - padding)
    ctx.lineTo(centerX - half + padding, centerY + half - padding)
    ctx.lineTo(centerX - half + padding, centerY + half - padding - len)
    ctx.stroke()
  }

  private drawChesses() {
    this.chesses.clearRect(0, 0, this.width, this.height)
    this.drawHints()
    this.board.forEach((row, _) => {
      Object.values(row).forEach((piece) => {
        piece.draw(this.gridSize)
      })
    })

    // Draw last move indicators
    if (this.moveHistory.length > 0) {
      const lastMove = this.moveHistory[this.moveHistory.length - 1]
      this.drawSelectionBox(lastMove.from.x, lastMove.from.y)
      this.drawSelectionBox(lastMove.to.x, lastMove.to.y)
    }
  }

  private drawBoard() {
    this.background.clearRect(0, 0, this.width, this.height)
    
    // Set background color (Wood texture color)
    this.background.fillStyle = '#eecfa1'
    this.background.fillRect(0, 0, this.width, this.height)

    const offsetX = this.gridSize / 2
    const offsetY = this.gridSize / 2

    // Draw thick border around the grid
    this.background.strokeStyle = '#5d4037' // Dark brown
    this.background.lineWidth = 4
    this.background.strokeRect(
      offsetX - 5,
      offsetY - 5,
      8 * this.gridSize + 10,
      9 * this.gridSize + 10
    )

    this.background.strokeStyle = '#5d4037' // Dark brown for grid lines
    this.background.lineWidth = 1.5

    const lineDrawer = Drawer.drawLine.bind(null, this.background)

    // Draw horizontal lines
    for (let i = 0; i < 10; i++) {
      lineDrawer(
        offsetX,
        offsetY + i * this.gridSize,
        offsetX + 8 * this.gridSize,
        offsetY + i * this.gridSize,
      )
    }

    // Draw vertical lines - but not across the river
    for (let i = 0; i < 9; i++) {
      lineDrawer(
        offsetX + i * this.gridSize,
        offsetY,
        offsetX + i * this.gridSize,
        offsetY + 4 * this.gridSize,
      )

      lineDrawer(
        offsetX + i * this.gridSize,
        offsetY + 5 * this.gridSize,
        offsetX + i * this.gridSize,
        offsetY + 9 * this.gridSize,
      )
    }

    // River text
    this.background.font = '30px KaiTi, SimHei, Arial'
    this.background.fillStyle = '#5d4037'
    this.background.textAlign = 'center'
    this.background.textBaseline = 'middle'
    
    // Rotate text for traditional look if desired, but simple placement first
    this.background.save()
    this.background.translate(offsetX + 2 * this.gridSize, offsetY + 4.5 * this.gridSize)
    // this.background.rotate(-Math.PI / 2) // Optional rotation
    this.background.fillText('楚 河', 0, 0)
    this.background.restore()

    this.background.save()
    this.background.translate(offsetX + 6 * this.gridSize, offsetY + 4.5 * this.gridSize)
    // this.background.rotate(Math.PI / 2) // Optional rotation
    this.background.fillText('汉 界', 0, 0)
    this.background.restore()

    // Draw the palaces (九宫)
    // Top palace
    lineDrawer(
      offsetX + 3 * this.gridSize,
      offsetY,
      offsetX + 5 * this.gridSize,
      offsetY + 2 * this.gridSize,
    )

    lineDrawer(
      offsetX + 5 * this.gridSize,
      offsetY,
      offsetX + 3 * this.gridSize,
      offsetY + 2 * this.gridSize,
    )

    lineDrawer(
      offsetX + 3 * this.gridSize,
      offsetY + 7 * this.gridSize,
      offsetX + 5 * this.gridSize,
      offsetY + 9 * this.gridSize,
    )

    lineDrawer(
      offsetX + 5 * this.gridSize,
      offsetY + 7 * this.gridSize,
      offsetX + 3 * this.gridSize,
      offsetY + 9 * this.gridSize,
    )

    // Vertical lines for river sides
    lineDrawer(
      offsetX,
      offsetY + 4 * this.gridSize,
      offsetX,
      offsetY + 5 * this.gridSize,
    )

    lineDrawer(
      offsetX + 8 * this.gridSize,
      offsetY + 4 * this.gridSize,
      offsetX + 8 * this.gridSize,
      offsetY + 5 * this.gridSize,
    )

    // Draw the position markers for soldiers/pawns
    this.drawPositionMarker(0, 3, offsetX, offsetY)
    this.drawPositionMarker(2, 3, offsetX, offsetY)
    this.drawPositionMarker(4, 3, offsetX, offsetY)
    this.drawPositionMarker(6, 3, offsetX, offsetY)
    this.drawPositionMarker(8, 3, offsetX, offsetY)

    this.drawPositionMarker(0, 6, offsetX, offsetY)
    this.drawPositionMarker(2, 6, offsetX, offsetY)
    this.drawPositionMarker(4, 6, offsetX, offsetY)
    this.drawPositionMarker(6, 6, offsetX, offsetY)
    this.drawPositionMarker(8, 6, offsetX, offsetY)

    // Draw the position markers for cannons
    this.drawPositionMarker(1, 2, offsetX, offsetY)
    this.drawPositionMarker(7, 2, offsetX, offsetY)
    this.drawPositionMarker(1, 7, offsetX, offsetY)
    this.drawPositionMarker(7, 7, offsetX, offsetY)
  }

  private drawPositionMarker(x: number, y: number, offsetX: number, offsetY: number) {
    const markerSize = 5

    const targetX = offsetX + x * this.gridSize
    const targetY = offsetY + y * this.gridSize

    const lineDrawer = Drawer.drawLine.bind(null, this.background)

    // Draw the position markers (small lines at corners)
    // Top-left
    if (x > 0) {
      lineDrawer(
        targetX - markerSize,
        targetY - markerSize,
        targetX - markerSize * 2,
        targetY - markerSize,
      )

      lineDrawer(
        targetX - markerSize,
        targetY - markerSize,
        targetX - markerSize,
        targetY - markerSize * 2,
      )
    }

    // Top-right
    if (x < 8) {
      lineDrawer(
        targetX + markerSize,
        targetY - markerSize,
        targetX + markerSize * 2,
        targetY - markerSize,
      )

      lineDrawer(
        targetX + markerSize,
        targetY - markerSize,
        targetX + markerSize,
        targetY - markerSize * 2,
      )
    }

    // Bottom-left
    if (x > 0) {
      lineDrawer(
        targetX - markerSize,
        targetY + markerSize,
        targetX - markerSize * 2,
        targetY + markerSize,
      )

      lineDrawer(
        targetX - markerSize,
        targetY + markerSize,
        targetX - markerSize,
        targetY + markerSize * 2,
      )
    }

    // Bottom-right
    if (x < 8) {
      lineDrawer(
        targetX + markerSize,
        targetY + markerSize,
        targetX + markerSize * 2,
        targetY + markerSize,
      )

      lineDrawer(
        targetX + markerSize,
        targetY + markerSize,
        targetX + markerSize,
        targetY + markerSize * 2,
      )
    }
  }

  /**
   * 公开方法：通过坐标执行移动（用于 AI）
   */
  public makeMoveByPosition(from: ChessPosition, to: ChessPosition) {
    this.move(from, to)
  }

  /**
   * 公开方法：获取棋盘（用于 AI 适配器）
   */
  public getBoard(): Board {
    return this.board
  }

  // 依据自定义棋子列表重建棋盘（用于残局初始化）
  private rebuildBoardFromPieces(pieces: CustomPieceDefinition[]) {
    this.board = Array.from({ length: 9 }).map(() => ({})) as Board
    let id = 0
    pieces.forEach((piece) => {
      const role: ChessRole = piece.color === this.selfColor ? 'self' : 'enemy'
      const chessPiece = ChessFactory.createChessPiece(
        this.chesses,
        id++,
        piece.type as any,
        piece.color,
        role,
        piece.position.x,
        this.gridSize,
      )
      chessPiece.position = { ...piece.position }
      this.board[piece.position.x][piece.position.y] = chessPiece
    })
  }

  // 静态方法：校验自定义布局的合法性
  public static validateCustomLayout(pieces: CustomPieceDefinition[]) {
    const errors: string[] = []
    const seen = new Set<string>()
    let redKing = 0
    let blackKing = 0
    const board: (CustomPieceDefinition | null)[][] = Array.from({ length: 9 }).map(() => Array(10).fill(null))
    pieces.forEach((piece) => {
      const { x, y } = piece.position
      if (x < 0 || x > 8 || y < 0 || y > 9) {
        errors.push(`棋子越界: ${piece.type}@${x},${y}`)
        return
      }
      const key = `${x},${y}`
      if (seen.has(key)) {
        errors.push(`位置重复: ${key}`)
        return
      }
      seen.add(key)
      if (piece.type === 'King') {
        const inPalace = x >= 3 && x <= 5 && (piece.color === 'red' ? y >= 7 && y <= 9 : y >= 0 && y <= 2)
        if (!inPalace) errors.push('将/帅未在九宫')
        piece.color === 'red' ? redKing++ : blackKing++
      }
      if (piece.type === 'Advisor') {
        const inPalace = x >= 3 && x <= 5 && (piece.color === 'red' ? y >= 7 && y <= 9 : y >= 0 && y <= 2)
        if (!inPalace) errors.push('士/仕未在九宫')
      }
      if (piece.type === 'Bishop') {
        const ownHalf = piece.color === 'red' ? y >= 5 : y <= 4
        if (!ownHalf) errors.push('象/相越河')
      }
      board[x][y] = piece
    })

    if (redKing !== 1 || blackKing !== 1) errors.push('需要红帅与黑将各一枚')

    // 将帅照面检测
    for (let x = 0; x < 9; x++) {
      let redY: number | null = null
      let blackY: number | null = null
      for (let y = 0; y < 10; y++) {
        const piece = board[x][y]
        if (piece && piece.type === 'King') {
          if (piece.color === 'red') redY = y
          else blackY = y
        }
      }
      if (redY !== null && blackY !== null) {
        let blocked = false
        for (let y = Math.min(redY, blackY) + 1; y < Math.max(redY, blackY); y++) {
          if (board[x][y]) {
            blocked = true
            break
          }
        }
        if (!blocked) errors.push('将帅照面')
      }
    }

    return { valid: errors.length === 0, errors }
  }

  // 将自定义棋子列表转为 engine 所需的 board 矩阵
  public static toEngineBoardFromPieces(pieces: CustomPieceDefinition[]) {
    const board = Array.from({ length: 10 }, () => Array(9).fill(null))
    const map: Record<CustomPieceType, string> = { King: 'k', Advisor: 'a', Bishop: 'b', Horse: 'n', Rook: 'r', Cannon: 'c', Pawn: 's' }
    pieces.forEach((p) => {
      board[p.position.y][p.position.x] = { t: map[p.type], c: p.color }
    })
    return board
  }
}

export default ChessBoard
