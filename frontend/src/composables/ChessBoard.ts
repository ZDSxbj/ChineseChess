import type { Board, ChessColor, ChessPiece, ChessPosition, ChessRole } from './ChessPiece'
import type { GameState } from '@/store/gameStore';
import { showMsg } from '@/components/MessageBox'
import { saveGameState, getGameState, clearGameState } from '@/store/gameStore'
import channel from '@/utils/channel'
import { ChessFactory, King } from './ChessPiece'
import Drawer from './drawer'

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
  private clickCallback: (event: MouseEvent) => void = () => {}
  // 新增：存储走棋历史（记录移动前的状态）
  private moveHistory: Array<{
    from: ChessPosition
    to: ChessPosition
    capturedPiece: ChessPiece | null// 被吃掉的棋子（如果有）
    currentRole: ChessRole // 记录当前回合角色，用于悔棋后恢复
  }> = []
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
  private isCheckmate(color: ChessColor): boolean {
    return this.isInCheck(color) && !this.canEscapeCheck(color)
  }

  // 检查当前回合方是否被将死，如果是则结束游戏
  private checkCurrentTurnCheckmate() {
    // 联机模式下由服务端/对手走子触发判定，本地模式才需要主动检测
    if (this.isNetPlay) return false
    // 确定当前回合应该走棋的颜色
    const enemyColor = this.selfColor === 'red' ? 'black' : 'red'
    const currentColor = this.currentRole === 'self' ? this.selfColor : enemyColor
    
    if (this.isCheckmate(currentColor)) {
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
      if (!piece || piece.color !== this.selectedPiece.color) {
        const curPiece = this.selectedPiece
        this.move(curPiece.position, { x, y })
        this.selectedPiece.deselect()
        this.selectedPiece = null
        return
      }
    }

    if (piece) {
      // 不是当前角色的棋子不能选中
      if (piece.color !== (this.currentRole === 'self' ? selfColor : enemyColor)) {
        return
      }
      // 不是自己的回合不能选中
      if (this.isNetPlay) {
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
    // 记录历史
    this.moveHistory.push({
      from: { ...from },
      to: { ...to },
      capturedPiece: targetPiece || null,
      currentRole: this.currentRole, // 记录当前角色，悔棋后需恢复
    })

    // 先在数据结构上应用移动（模拟新的棋盘状态），再绘制
    delete this.board[from.x][from.y]
    this.board[to.x][to.y] = piece
    piece.move(to)
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

    // 检查对方是否被将军或将死
    const opponentColor = piece.color === 'red' ? 'black' : 'red'
    const opponentInCheck = this.isInCheck(opponentColor)
    const opponentCheckmated = opponentInCheck && this.isCheckmate(opponentColor)

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
      // 对方被将军或我方被将军（取决于当前移动方），播放提示并响铃
      showMsg(`${opponentColor === 'red' ? '红方' : '黑方'}被将军！`)
      try {
        if (this.checkAudio) {
          this.checkAudio.currentTime = 0
          void this.checkAudio.play()
        }
      } catch {}
    } else {
      // 非将军的情况下，如果发生了吃子，则播放吃子音效（仅联机时启用）
      if (this.isNetPlay && !!targetPiece) {
        try {
          if (this.eatAudio) {
            this.eatAudio.currentTime = 0
            void this.eatAudio.play()
          }
        } catch {}
      }
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
    saveGameState({
      isNetPlay: this.isNetPlay,
      selfColor: this.selfColor,
      moveHistory: this.moveHistory,
      currentRole: this.currentRole,
    })
    return true
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

    this.selectedPiece?.deselect()
    this.selectedPiece = piece
    piece.select()
    // 选中棋子音效（仅联机时启用）
    if (this.isNetPlay) {
      this.selectAudio?.play().catch(() => {})
    }
  }

  public start(color: ChessColor, isNet: boolean) {
    this.isNetPlay = isNet
    this.selfColor = color
    this.currentRole = color === 'red' ? 'self' : 'enemy'
    this.board = Array.from({ length: 9 }).fill(null).map(() => {
      return {}
    })
    this.drawBoard()
    this.initChesses()
    this.drawChesses()
    this.listenClick()
    this.listenEvent()
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
  }

  private listenMove(req: { from: ChessPosition, to: ChessPosition }) {
    const { from, to } = req
    this.move(from, to)
  }

  private listenEvent() {
    channel.on('NET:CHESS:MOVE', this.listenMove.bind(this))
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

  private drawChesses() {
    this.chesses.clearRect(0, 0, this.width, this.height)
    this.board.forEach((row, _) => {
      Object.values(row).forEach((piece) => {
        piece.draw(this.gridSize)
      })
    })
  }

  private drawBoard() {
    this.background.clearRect(0, 0, this.width, this.height)
    this.background.strokeStyle = '#000'
    this.background.lineWidth = 2

    // Set background color
    this.background.fillStyle = '#f4d1a4'
    this.background.fillRect(0, 0, this.width, this.height)

    const offsetX = this.gridSize / 2
    const offsetY = this.gridSize / 2

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

    this.background.font = '20px Arial'
    this.background.fillStyle = '#000'
    this.background.textAlign = 'center'
    this.background.fillText('楚 河', offsetX + 2 * this.gridSize, offsetY + 4.5 * this.gridSize)
    this.background.fillText('汉 界', offsetX + 6 * this.gridSize, offsetY + 4.5 * this.gridSize)

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
}

export default ChessBoard
