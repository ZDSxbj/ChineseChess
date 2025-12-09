package user

import (
	"fmt"
	"time"
)

// SaveGameRecordRequest 用于接收前端提交的人机对战对局记录
type SaveGameRecordRequest struct {
	// 提交者是否为红方
	IsRed bool `json:"is_red"`
	// 提交者视角的结果: 0=胜,1=负,2=和
	Result int `json:"result"`
	// 棋局历史，按项目中使用为字符串格式（例如 SAN 或自定义格式）
	History string `json:"history"`
	// 对局开始时间，可选
	StartTime time.Time `json:"start_time"`
}

func (r *SaveGameRecordRequest) Examine() error {
	if r.Result < 0 || r.Result > 2 {
		return fmt.Errorf("无效的结果值")
	}
	if r.History == "" {
		return fmt.Errorf("历史不能为空")
	}
	return nil
}
