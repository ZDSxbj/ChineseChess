package websocket

type CommendType int

const (
	commandRegister       CommendType = iota + 1 // 注册
	commandUnregister                            // 注销
	commandMatch                                 // 匹配
	commandMove                                  // 移动
	commandSendMessage                           // 发送消息
	commandStart                                 // 开始游戏
	commandEnd                                   // 结束游戏
	commandJoin                                  // 加入房间
	commandCreate                                // 创建房间
	commandHeartbeat                             // 心跳
	commandRegretRequest  CommendType = 13       // 悔棋请求命令
	commandRegretResponse CommendType = 14       // 悔棋响应命令
	commandDrawRequest    CommendType = 15       // 和棋请求命令
	commandDrawResponse   CommendType = 16       // 和棋响应命令
)

type moveRequest struct {
	from *Client
	move MoveMessage
}

type sendMessageRequest struct {
	target  *Client
	message any
}

type hubCommand struct {
	commandType CommendType
	client      *Client
	payload     any
}

// 新增：悔棋请求 payload 结构（用于在命令中传递数据）
type regretRequestPayload struct {
	from *Client // 发起悔棋的客户端
}

// 新增：悔棋响应 payload 结构
type regretResponsePayload struct {
	from     *Client // 响应悔棋的客户端
	accepted bool    // 是否同意悔棋
}

type drawRequestPayload struct {
	from *Client // 发起和棋的客户端
}

// 新增：和棋响应 payload 结构
type drawResponsePayload struct {
	from     *Client // 响应和棋的客户端
	accepted bool    // 是否同意和棋
}
