interface UserInfo {
  id?: number
  name: string
  avatar?: string
  gender?: string
  exp?: number
  onlineStatus?: '在线' | '离线'
  totalGames?: number
  winGames?: number
  loseGames?: number
  winRate?: number
  email?: string
}
