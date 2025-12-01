export interface UserInfo {
  id: number
  name: string
  email: string      // 邮箱
  avatar: string  
  gender: string  
  exp: number
  onlineStatus: '在线' | '离线'; 
  totalGames: number;
  winGames: number;
  loseGames: number;
  winRate: number;
}

