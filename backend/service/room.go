// package service

// import (
// 	"chinese-chess-backend/database"
// 	"chinese-chess-backend/dto/room"
// 	userModel "chinese-chess-backend/model/user"
// )

// type RoomService struct{}

// func NewRoomService() *RoomService {
// 	return &RoomService{}
// }

// func (rs *RoomService) GetSpareRooms(req room.GetSpareRoomsRequest) (room.GetSpareRoomsResponse, error) {
//     db := database.GetMysqlDb()
// 	rooms := req.Infos
// 	var resp room.GetSpareRoomsResponse

//     // Collect all user IDs that need to be fetched
//     var userIDs []uint
//     userIDPositionMap := make(map[uint][]struct{
//         roomIndex int
//         isCurrentUser bool
//     })

//     for i, room := range rooms {
//         if room.Current.ID != 0 {
//             id := uint(room.Current.ID)
//             userIDs = append(userIDs, id)
//             userIDPositionMap[id] = append(userIDPositionMap[id], struct{
//                 roomIndex int
//                 isCurrentUser bool
//             }{i, true})
//         }

//         if room.Next.ID != 0 {
//             id := uint(room.Next.ID)
//             userIDs = append(userIDs, id)
//             userIDPositionMap[id] = append(userIDPositionMap[id], struct{
//                 roomIndex int
//                 isCurrentUser bool
//             }{i, false})
//         }
//     }

//     // If no users to fetch, return the original array
//     if len(userIDs) == 0 {
//         return resp, nil
//     }

//     // Fetch all users at once with only the needed fields
//     var users []userModel.User
//     if err := db.Model(&userModel.User{}).
//         Select("id, name, exp").
//         Where("id IN ?", userIDs).
//         Find(&users).Error; err != nil {
//         return resp, err
//     }

//     // Create a map for quick lookup
//     userMap := make(map[uint]userModel.User)
//     for _, user := range users {
//         userMap[user.ID] = user
//     }

//     // Update rooms with user information
//     for _, user := range users {
//         positions := userIDPositionMap[user.ID]
//         for _, pos := range positions {
//             if pos.isCurrentUser {
//                 rooms[pos.roomIndex].Current.Name = user.Name
//                 rooms[pos.roomIndex].Current.Exp = user.Exp
//             } else {
//                 rooms[pos.roomIndex].Next.Name = user.Name
//                 rooms[pos.roomIndex].Next.Exp = user.Exp
//             }
//         }
//     }

// 	resp.Rooms = rooms
// 	return resp, nil
// }

package service

import (
	"fmt"

	"chinese-chess-backend/database"
	"chinese-chess-backend/dto/room"
	"chinese-chess-backend/dto/user"
	roomModel "chinese-chess-backend/model/room"
	userModel "chinese-chess-backend/model/user"
)

type RoomService struct{}

func NewRoomService() *RoomService {
	return &RoomService{}
}

// GetSpareRooms 获取空闲房间（等待中的房间）
func (rs *RoomService) GetSpareRooms(req room.GetSpareRoomsRequest) (room.GetSpareRoomsResponse, error) {
	db := database.GetMysqlDb()
	var resp room.GetSpareRoomsResponse

	// 设置分页默认值
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	// 查询空闲房间
	var dbRooms []roomModel.Room
	var total int64

	// 计算总数
	db.Model(&roomModel.Room{}).Where("status = ?", "waiting").Count(&total)

	// 查询数据（预加载用户信息）
	if err := db.Preload("Player1").Preload("Player2").
		Where("status = ?", "waiting").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&dbRooms).Error; err != nil {
		return resp, err
	}

	// 转换为DTO
	rooms := make([]room.RoomDTO, 0, len(dbRooms))
	for _, dbRoom := range dbRooms {
		roomDTO := room.RoomDTO{
			ID:     dbRoom.ID,
			Name:   dbRoom.Name,
			Status: dbRoom.Status,
			Current: user.UserInfo{
				ID:   dbRoom.Player1.ID,
				Name: dbRoom.Player1.Name,
				Exp:  dbRoom.Player1.Exp,
			},
		}

		// 如果有第二个玩家
		if dbRoom.Player2ID != 0 {
			roomDTO.Next = user.UserInfo{
				ID:   dbRoom.Player2.ID,
				Name: dbRoom.Player2.Name,
				Exp:  dbRoom.Player2.Exp,
			}
		}

		rooms = append(rooms, roomDTO)
	}

	resp.Rooms = rooms
	resp.Total = total
	return resp, nil
}

// CreateRoom 创建房间
func (rs *RoomService) CreateRoom(playerID uint, roomName string) (room.RoomDTO, error) {
	db := database.GetMysqlDb()

	// 首先获取玩家信息
	var player userModel.User
	if err := db.First(&player, playerID).Error; err != nil {
		return room.RoomDTO{}, fmt.Errorf("用户不存在")
	}

	// 创建房间
	newRoom := roomModel.Room{
		Name:      roomName,
		Player1ID: playerID,
		Status:    "waiting",
		Board:     `{"fen": "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"}`,
	}

	if err := db.Create(&newRoom).Error; err != nil {
		return room.RoomDTO{}, fmt.Errorf("创建房间失败: %v", err)
	}

	// 返回创建的房间信息
	return room.RoomDTO{
		ID:     newRoom.ID,
		Name:   newRoom.Name,
		Status: newRoom.Status,
		Current: user.UserInfo{
			ID:   player.ID,
			Name: player.Name,
			Exp:  player.Exp,
		},
	}, nil
}

// JoinRoom 加入房间
func (rs *RoomService) JoinRoom(roomID uint, playerID uint) error {
	db := database.GetMysqlDb()

	// 检查房间是否存在且可加入
	var dbRoom roomModel.Room
	if err := db.First(&dbRoom, roomID).Error; err != nil {
		return fmt.Errorf("房间不存在")
	}

	if dbRoom.Status != "waiting" {
		return fmt.Errorf("房间不可加入")
	}

	if dbRoom.Player2ID != 0 {
		return fmt.Errorf("房间已满")
	}

	if dbRoom.Player1ID == playerID {
		return fmt.Errorf("不能加入自己创建的房间")
	}

	// 更新房间的第二个玩家
	if err := db.Model(&roomModel.Room{}).
		Where("id = ?", roomID).
		Update("player2_id", playerID).Error; err != nil {
		return fmt.Errorf("加入房间失败: %v", err)
	}

	return nil
}

// GetRoomInfo 获取房间详细信息
func (rs *RoomService) GetRoomInfo(roomID uint) (room.RoomDTO, error) {
	db := database.GetMysqlDb()
	var dbRoom roomModel.Room

	if err := db.Preload("Player1").Preload("Player2").
		First(&dbRoom, roomID).Error; err != nil {
		return room.RoomDTO{}, fmt.Errorf("房间不存在")
	}

	// 转换为DTO
	roomDTO := room.RoomDTO{
		ID:     dbRoom.ID,
		Name:   dbRoom.Name,
		Status: dbRoom.Status,
		Board:  dbRoom.Board,
		Current: user.UserInfo{
			ID:   dbRoom.Player1.ID,
			Name: dbRoom.Player1.Name,
			Exp:  dbRoom.Player1.Exp,
		},
	}

	// 如果有第二个玩家
	if dbRoom.Player2ID != 0 {
		roomDTO.Next = user.UserInfo{
			ID:   dbRoom.Player2.ID,
			Name: dbRoom.Player2.Name,
			Exp:  dbRoom.Player2.Exp,
		}
	}

	return roomDTO, nil
}

// LeaveRoom 离开房间
func (rs *RoomService) LeaveRoom(roomID uint, playerID uint) error {
	db := database.GetMysqlDb()

	// 查询房间信息
	var dbRoom roomModel.Room
	if err := db.First(&dbRoom, roomID).Error; err != nil {
		return fmt.Errorf("房间不存在")
	}

	// 判断玩家在房间中的位置并更新
	if dbRoom.Player1ID == playerID {
		// 房主离开，如果房间有第二个玩家，则第二个玩家成为房主
		if dbRoom.Player2ID != 0 {
			if err := db.Model(&dbRoom).
				Updates(map[string]interface{}{
					"player1_id": dbRoom.Player2ID,
					"player2_id": 0,
				}).Error; err != nil {
				return fmt.Errorf("离开房间失败: %v", err)
			}
		} else {
			// 没有其他玩家，直接删除房间
			if err := db.Delete(&dbRoom).Error; err != nil {
				return fmt.Errorf("离开房间失败: %v", err)
			}
		}
	} else if dbRoom.Player2ID == playerID {
		// 第二位玩家离开，清空位置
		if err := db.Model(&dbRoom).Update("player2_id", 0).Error; err != nil {
			return fmt.Errorf("离开房间失败: %v", err)
		}
	} else {
		return fmt.Errorf("玩家不在该房间中")
	}

	return nil
}

// StartGame 开始游戏
func (rs *RoomService) StartGame(roomID uint, playerID uint) error {
	db := database.GetMysqlDb()

	// 查询房间信息
	var dbRoom roomModel.Room
	if err := db.First(&dbRoom, roomID).Error; err != nil {
		return fmt.Errorf("房间不存在")
	}

	// 检查权限（只有房主可以开始游戏）
	if dbRoom.Player1ID != playerID {
		return fmt.Errorf("只有房主可以开始游戏")
	}

	// 检查房间状态
	if !dbRoom.CanStart() {
		return fmt.Errorf("房间尚未满员，无法开始游戏")
	}

	// 更新房间状态为游戏中
	if err := db.Model(&dbRoom).Update("status", "playing").Error; err != nil {
		return fmt.Errorf("开始游戏失败: %v", err)
	}

	return nil
}
