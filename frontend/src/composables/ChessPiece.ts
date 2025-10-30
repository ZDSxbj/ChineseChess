// 車 馬 象 士 炮 兵 將
type ChessNames = ['Rook', 'Horse', 'Bishop', 'Advisor', 'Cannon', 'Pawn', 'King']
export type ChessColor = 'red' | 'black'
export type ChessRole = 'self' | 'enemy'
export type Board = Array<{ [key: string]: ChessPiece }>
export interface ChessPosition { x: number, y: number }

class ChessPiece {
  public id: number
  public name: string
  public color: ChessColor
  public role: ChessRole = 'enemy'
  public position: ChessPosition
  private gridSize: number = 50 // 棋盘格子大小
  public isSelected: boolean
  private ctx: CanvasRenderingContext2D
  private flashingInterval: number = 0
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    name: string,
    color: ChessColor,
    role: ChessRole,
    position: ChessPosition,
    gridSize: number,
  ) {
    this.ctx = ctx
    this.id = id
    this.name = name
    this.color = color
    this.role = role
    this.position = position
    this.isSelected = false
    this.gridSize = gridSize
  }

  get radius() {
    return this.gridSize / 2
  }

  public select() {
    if (this.isSelected)
      return
    this.isSelected = true
    // 闪烁效果
    let count = 0
    this.flashingInterval = setInterval(() => {
      count++
      count & 1 ? this.clearFromCanvas() : this.draw()
    }, 250)
  }

  public deselect() {
    this.isSelected = false
    // 取消闪烁效果
    clearInterval(this.flashingInterval)
    this.flashingInterval = 0
    this.clearFromCanvas()
    this.draw()
  }

  public draw(size: number = this.gridSize) {
    this.gridSize = size
    const ctx = this.ctx
    const x = this.position.x * this.gridSize + this.gridSize / 2
    const y = this.position.y * this.gridSize + this.gridSize / 2
    const pieceSize = this.gridSize * 1.1 // 棋子图片大小（略小于格子，留边距）

    // 关键：建立棋子汉字名称到英文文件名的映射表
    const nameMap: Record<string, string> = {
      俥: 'rook', // 红方车
      車: 'rook', // 黑方车
      傌: 'horse', // 红方马
      馬: 'horse', // 黑方马
      相: 'bishop', // 红方象
      象: 'bishop', // 黑方象
      仕: 'advisor', // 黑方士（敌方）
      士: 'advisor', // 红方士（己方）
      炮: 'cannon', // 炮（红黑通用名称）
      兵: 'pawn', // 红方兵
      卒: 'pawn', // 黑方卒
      帥: 'king', // 红方帅
      將: 'king', // 黑方将
    }

    // 根据当前棋子的汉字名称获取对应的英文文件名
    const englishName = nameMap[this.name]
    if (!englishName) {
      console.error(`未找到棋子${this.name}的映射关系`)
      this.drawFallback(x, y) // 降级绘制
      return
    }
    // 生成正确的图片文件名（例如 red_horse.png）
    const imageName = `${this.color}_${englishName}.png`
    const img = new Image()
    img.src = `/chess/images/${imageName}`// 修正后的路径 // public目录下的路径

    // 图片加载完成后绘制
    img.onload = () => {
      // 清除原有位置（避免残影）
      this.clearFromCanvas()
      // 绘制图片（居中显示）
      ctx.drawImage(
        img,
        x - pieceSize / 2, // 图片左上角x
        y - pieceSize / 2, // 图片左上角y
        pieceSize, // 图片宽度
        pieceSize, // 图片高度
      )
    }
  }

  private drawFallback(x: number, y: number) {
    // 保留原有的圆形+文字绘制逻辑
    this.ctx.beginPath()
    this.ctx.arc(x, y, this.radius, 0, Math.PI * 2)
    this.ctx.fillStyle = this.color === 'red' ? '#f44336' : '#212121'
    this.ctx.fill()
    this.ctx.fillStyle = '#ffffff'
    this.ctx.font = 'bold 20px SimHei, Arial'
    this.ctx.textAlign = 'center'
    this.ctx.textBaseline = 'middle'
    this.ctx.fillText(this.name, x, y)
  }

  private clearFromCanvas() {
    const ctx = this.ctx
    const x = this.position.x * this.gridSize + this.gridSize / 2
    const y = this.position.y * this.gridSize + this.gridSize / 2

    const clearRadius = this.radius
    ctx.clearRect(x - clearRadius, y - clearRadius, clearRadius * 2, clearRadius * 2)
  }

  // 坐标由棋盘处理，这里接收的是处理好的坐标
  // 这里的坐标是棋盘坐标系，0-8,0-9
  public move(newPosition: ChessPosition) {
    // 清除原来位置
    this.clearFromCanvas()
    // 更新位置
    this.position = newPosition
    // 绘制新位置
    this.draw()
    return true
  }

  public isMoveValid(newPosition: ChessPosition): boolean {
    const { x, y } = newPosition
    if (x < 0 || x > 8 || y < 0 || y > 9) {
      return false
    }
    if (x === this.position.x && y === this.position.y) {
      return false
    }
    return true
  }
}

class King extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    gridSize: number = 50,
  ) {
    const name = color === 'red' ? '帥' : '將'
    const position = role === 'enemy' ? { x: 4, y: 0 } : { x: 4, y: 9 }
    super(ctx, id, name, color, role, position, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition, board?: Board): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    if (!board)
      throw new Error('Board is required for King movement validation')
    const { x, y } = newPosition
    if (x < 3 || x > 5) {
      return false
    }

    let nums = 0
    // 检查路径上是否有棋子
    for (let i = Math.min(this.position.y, y) + 1; i < Math.max(this.position.y, y); i++) {
      if (board[this.position.x][i]) {
        nums++
      }
    }
    if (nums === 0) {
      const targetPiece = board[x][y]
      if (targetPiece && targetPiece instanceof King) {
        return true
      }
    }

    const { upY, downY } = this.role === 'enemy' ? { upY: 0, downY: 2 } : { upY: 7, downY: 9 }
    if (y < upY || y > downY) {
      return false
    }

    if ((x - this.position.x) ** 2 + (y - this.position.y) ** 2 > 2) {
      return false
    }

    return true
  }
}

class Advisor extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ) {
    const name = role === 'enemy' ? '仕' : '士'
    const y = role === 'enemy' ? 0 : 9
    super(ctx, id, name, color, role, { x, y }, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    const { x, y } = newPosition

    if (x < 3 || x > 5) {
      return false
    }

    const { upY, downY } = this.role === 'enemy' ? { upY: 0, downY: 2 } : { upY: 7, downY: 9 }
    if (y < upY || y > downY) {
      return false
    }

    if (Math.abs(x - this.position.x) + Math.abs(y - this.position.y) !== 2) {
      return false
    }

    return true
  }
}

class Bishop extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ) {
    const name = color === 'red' ? '相' : '象'
    const y = role === 'enemy' ? 0 : 9
    super(ctx, id, name, color, role, { x, y }, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition, board?: Board): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    if (!board)
      throw new Error('Board is required for King movement validation')
    const { x, y } = newPosition

    if (Math.abs(x - this.position.x) !== 2 || Math.abs(y - this.position.y) !== 2) {
      return false
    }

    const { upY, downY } = this.role === 'enemy' ? { upY: 0, downY: 4 } : { upY: 5, downY: 9 }
    if (y < upY || y > downY) {
      return false
    }

    const midX = (this.position.x + x) / 2
    const midY = (this.position.y + y) / 2
    if (board[midX][midY]) {
      return false
    }
    return true
  }
}

class Pawn extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ) {
    const name = color === 'red' ? '兵' : '卒'
    const y = role === 'enemy' ? 3 : 6
    super(ctx, id, name, color, role, { x, y }, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    const { x, y } = newPosition

    if (Math.abs(x - this.position.x) + Math.abs(y - this.position.y) !== 1) {
      return false
    }

    const river = this.role === 'enemy' ? 4 : 5
    if (this.role === 'enemy') {
      if (y - this.position.y < 0) {
        return false
      }
      if (this.position.y <= river) {
        if (x !== this.position.x) {
          return false
        }
        return true
      }

      // 过河了
      return true
    }
    else {
      if (this.position.y - y < 0) {
        return false
      }
      if (this.position.y >= river) {
        if (x !== this.position.x) {
          return false
        }
        return true
      }

      // 过河了
      return true
    }
  }
}

class Rook extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ) {
    const name = color === 'red' ? '俥' : '車'
    const y = role === 'enemy' ? 0 : 9
    super(ctx, id, name, color, role, { x, y }, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition, board?: Board): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    if (!board)
      throw new Error('Board is required for King movement validation')
    // const arr: ChessPosition[] = []
    const { x, y } = newPosition
    if (x !== this.position.x && y !== this.position.y) {
      return false
    }
    // 检查路径上是否有棋子
    let has = false
    if (x === this.position.x) {
      for (let i = Math.min(this.position.y, y) + 1; i < Math.max(this.position.y, y); i++) {
        if (board[this.position.x][i]) {
          has = true
          break
        }
      }
    }
    else {
      for (let i = Math.min(this.position.x, x) + 1; i < Math.max(this.position.x, x); i++) {
        if (board[i][this.position.y]) {
          has = true
          break
        }
      }
    }
    if (has) {
      return false
    }

    return true
  }
}

class Horse extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ) {
    const name = color === 'red' ? '傌' : '馬'
    const y = role === 'enemy' ? 0 : 9
    super(ctx, id, name, color, role, { x, y }, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition, board?: Board): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    if (!board)
      throw new Error('Board is required for King movement validation')
    const { x, y } = newPosition
    const dx = Math.abs(x - this.position.x)
    const dy = Math.abs(y - this.position.y)
    if (dx + dy !== 3 || dx === 0 || dy === 0) {
      return false
    }

    // 检查是否有障碍物
    const direction = dx > dy
    let checkPosition
    if (direction) {
      const directionX = x - this.position.x > 0 ? 1 : -1
      checkPosition = { x: this.position.x + directionX, y: this.position.y }
    }
    else {
      const directionY = y - this.position.y > 0 ? 1 : -1
      checkPosition = { x: this.position.x, y: this.position.y + directionY }
    }
    if (board[checkPosition.x][checkPosition.y]) {
      return false
    }

    return true
  }
}

// 炮
class Cannon extends ChessPiece {
  constructor(
    ctx: CanvasRenderingContext2D,
    id: number,
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ) {
    const name = '炮'
    const y = role === 'enemy' ? 2 : 7
    super(ctx, id, name, color, role, { x, y }, gridSize)
  }

  public isMoveValid(newPosition: ChessPosition, board?: Board): boolean {
    if (!super.isMoveValid(newPosition)) {
      return false
    }
    if (!board)
      throw new Error('Board is required for King movement validation')
    // const arr: ChessPosition[] = []
    const { x, y } = newPosition

    if (x !== this.position.x && y !== this.position.y) {
      return false
    }

    // 检查路径上是否有棋子
    let nums = 0
    if (x === this.position.x) {
      for (let i = Math.min(this.position.y, y) + 1; i < Math.max(this.position.y, y); i++) {
        if (board[this.position.x][i]) {
          nums++
        }
      }
    }
    else {
      for (let i = Math.min(this.position.x, x) + 1; i < Math.max(this.position.x, x); i++) {
        if (board[i][this.position.y]) {
          nums++
        }
      }
    }
    const piece = board[x][y]
    if (piece) {
      nums++
    }
    if (nums !== 2 && nums !== 0) {
      return false
    }

    return true
  }
}

class ChessFactory {
  public static createChessPiece(
    ctx: CanvasRenderingContext2D,
    id: number,
    name: ChessNames[number],
    color: ChessColor,
    role: ChessRole,
    x: number,
    gridSize: number = 50,
  ): ChessPiece {
    switch (name) {
      case 'King':
        return new King(ctx, id, color, role, gridSize)
      case 'Advisor':
        return new Advisor(ctx, id, color, role, x, gridSize)
      case 'Bishop':
        return new Bishop(ctx, id, color, role, x, gridSize)
      case 'Pawn':
        return new Pawn(ctx, id, color, role, x, gridSize)
      case 'Rook':
        return new Rook(ctx, id, color, role, x, gridSize)
      case 'Horse':
        return new Horse(ctx, id, color, role, x, gridSize)
      case 'Cannon':
        return new Cannon(ctx, id, color, role, x, gridSize)
      default:
        console.log(name)
        throw new Error('Invalid chess piece name')
    }
  }
}

export { ChessFactory, ChessPiece, King }
