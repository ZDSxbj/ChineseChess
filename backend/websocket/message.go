package websocket

type MessageType int

// 信息类型
const (
	messageNormal         MessageType = iota + 1 // 普通消息
	messageMatch                                 // 匹配消息
	messageMove                                  // 移动消息
	messageStart                                 // 开始消息
	messageEnd                                   // 结束消息
	messageJoin                                  // 加入消息
	messageCreate                                // 创建房间消息
	messageGiveUp                                // 放弃消息
	messageError                      = 10
	messageRegretRequest  MessageType = 11 // 悔棋请求
	messageRegretResponse MessageType = 12 // 悔棋响应
	messageDrawRequest    MessageType = 13 // 和棋请求
	messageDrawResponse   MessageType = 14 // 和棋响应
	messageChatMessage    MessageType = 15 // 聊天消息 (与前端保持一致)
	messageSync           MessageType = 16 // 同步当前房间状态（重连使用）
	messageFriendRequest  MessageType = 17 // 好友申请（通过 websocket 发送/推送）
)

type BaseMessage struct {
	Type MessageType `json:"type"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MoveMessage struct {
	BaseMessage
	From Position `json:"from"`
	To   Position `json:"to"`
}

type NormalMessage struct {
	BaseMessage
	Message string `json:"message"`
}

type startMessage struct {
	BaseMessage
	Role string `json:"role"`
}

type joinMessage struct {
	BaseMessage
	RoomId int `json:"roomId"`
}

type endMessage struct {
	BaseMessage
	Winner clientRole `json:"winner"`
}

type RegretResponseMessage struct {
	BaseMessage
	Accepted bool `json:"accepted"`
}

type DrawResponseMessage struct {
	BaseMessage
	Accepted bool `json:"accepted"`
}

type ChatMessage struct {
	BaseMessage
	Content    string `json:"content"`
	Sender     string `json:"sender"`
	RelationId uint   `json:"relationId,omitempty"`
	SenderId   uint   `json:"senderId,omitempty"`
	MessageId  uint   `json:"messageId,omitempty"`
	CreatedAt  int64  `json:"createdAt,omitempty"`
}

type FriendRequestMessage struct {
	BaseMessage
	RequestId  uint   `json:"requestId,omitempty"`
	SenderId   uint   `json:"senderId,omitempty"`
	ReceiverId uint   `json:"receiverId,omitempty"`
	Content    string `json:"content,omitempty"`
	CreatedAt  int64  `json:"createdAt,omitempty"`
	SenderName string `json:"senderName,omitempty"`
}

// SyncMessage 用于在玩家重连时将房间当前的棋步历史、玩家角色和当前轮次同步给客户端
type SyncMessage struct {
	BaseMessage
	History     []Position `json:"history"`
	Role        string     `json:"role"`
	CurrentTurn string     `json:"currentTurn"`
}
