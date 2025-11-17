// package room

// type GetSpareRoomsRequest struct {
// 	Infos []RoomInfo `json:"-"`
// }

// type GetSpareRoomsResponse struct {
// 	Rooms []RoomInfo `json:"rooms"`
// }

package room

// GetSpareRoomsRequest 获取空闲房间请求
type GetSpareRoomsRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// GetSpareRoomsResponse 获取空闲房间响应
type GetSpareRoomsResponse struct {
	Rooms []RoomDTO `json:"rooms"`
	Total int64     `json:"total"`
}
