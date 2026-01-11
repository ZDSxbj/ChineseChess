# 中国象棋对弈网站

一个支持实时对弈、好友系统、聊天与残局挑战的中国象棋网站。前端使用 Vue 3 + Vite，后端使用 Go (Gin + GORM + Redis + MySQL)，通过 WebSocket 实时同步棋局。项目内置 Nginx 与 Docker Compose，开箱即用。

## 功能特性
- 实时对弈：匹配/创建/加入房间、走子广播、认输、悔棋与和棋请求、断线重连同步
- 好友系统：申请/接受/拒绝/删除好友，好友对弈邀请与撤回、接受/拒绝
- 私聊功能：好友之间发送与拉取聊天消息，已读标记
- 残局挑战：进度保存、结算经验奖励
- 账户体系：注册/登录（邮箱验证码）、资料更新、头像上传、邮箱/密码修改、退出登录、在线心跳
- 对战记录：保存与查询历史对局
- AI 对弈：可创建 AI 房间（实验特性）

## 技术栈
- 前端：Vue 3、Vite、Pinia、UnoCSS，生产环境由 Nginx 托管静态资源
- 后端：Go、Gin、GORM、MySQL、Redis、Gorilla WebSocket
- 基础设施：Docker、Docker Compose、Nginx 反向代理

## 目录结构
- 后端：见 [backend](backend)（Gin 路由见 [backend/route/route.go](backend/route/route.go)）
- 前端：见 [frontend](frontend)（Vite 基础路径为 `/chess`，见 [frontend/vite.config.ts](frontend/vite.config.ts)）
- 反向代理：Nginx 路由前缀 `/chess`、API `/chess/api/`、WebSocket `/chess/ws`，见 [nginx.conf](nginx.conf)
- 编排：服务定义见 [docker-compose.yml](docker-compose.yml)

## 一键启动（Docker Compose）
适合快速体验与本地一体化环境。

1) 准备 SMTP 配置（用于邮箱验证码）
- 复制并编辑后端配置文件 [backend/config.json](backend/config.json)（容器内以只读挂载到 `/app/config.json`）。示例：

```json
{
	"smtp": {
		"host": "smtp.example.com",
		"port": "465",
		"username": "your_email@example.com",
		"password": "your_app_password"
	}
}
```

提示：请使用你自己的邮箱服务与应用专用密码。不要将真实凭据提交到仓库。

2) 启动服务

```bash
docker compose up -d
```

3) 访问入口
- 前端（同源反代）：http://localhost/chess
- API（同源反代）：http://localhost/chess/api
- WebSocket（同源反代）：ws://localhost/chess/ws

说明：在 Compose 场景下，前端通过同源路径 `/chess/api` 与 `/chess/ws` 访问后端，无需额外 CORS 配置。

4) 关闭与清理

```bash
docker compose down
```

数据持久化：
- MySQL 使用卷 `mysql-data`（见 [docker-compose.yml](docker-compose.yml)）
- 用户上传（头像等）持久化于 [backend/uploads](backend/uploads)

## 本地开发
如果你希望分别启动前后端进行开发调试，可按以下步骤：

### 前置依赖
- Node.js LTS（前端）
- Go（后端，建议与 [backend/go.mod](backend/go.mod) 对应版本或更高）
- 本地 MySQL 与 Redis（或使用 Docker 启动服务）

### 启动后端（本地）
1) 准备数据库与缓存
- MySQL：创建数据库 `chinese_chess`，并记录连接信息
- Redis：本地默认 `localhost:6379`

2) 配置环境变量（示例）

```bash
# 数据库
set DB_HOST=localhost
set DB_PORT=3306
set DB_USERNAME=root
set DB_PASSWORD=your_password
set DB_NAME=chinese_chess
set DB_LOG_LEVEL=silent

# Redis（如有密码可设置 REDIS_PASSWORD）
set REDIS_HOST=localhost
set REDIS_PORT=6379

# CORS 允许的前端来源（本地 Vite 默认端口）
set FRONTEND_URL=http://localhost:5173
```

3) 配置邮箱（与 Docker 一致）
- 确保 [backend/config.json](backend/config.json) 正确填写 SMTP 信息

4) 启动后端

```bash
cd backend
go mod download
go run main.go
```

后端默认监听 `:8080`，路由前缀为 `/api`，WebSocket 为 `/ws`。

### 启动前端（本地）
1) 配置前端环境变量（如使用本地后端）：编辑 [frontend/.env](frontend/.env) 或新建 `.env.local`：

```dotenv
VITE_API_URL=http://localhost:8080/api
VITE_WEBSOCKET_URL=ws://localhost:8080/ws
```

2) 安装依赖并启动

```bash
cd frontend
npm install
npm run dev
```

打开 http://localhost:5173 进行调试。

## 生产部署要点
- 路径前缀：本项目生产默认以 `/chess` 为前缀（见 [frontend/vite.config.ts](frontend/vite.config.ts) 与 [nginx.conf](nginx.conf)）。如需修改，请同时调整前端 `base`、Nginx 路由以及前端 API/WebSocket 访问路径。
- 反向代理：Nginx 已将静态资源映射到 `/chess`，API 代理到 `/chess/api/`，WebSocket 代理到 `/chess/ws`。
- CORS：同源部署无须 CORS。若前后端分离部署，请在后端通过环境变量 `FRONTEND_URL` 配置允许的来源，多个地址用逗号分隔。
- 持久化：保证 MySQL 卷与 `backend/uploads`（用户头像等）持久化存储。
- 邮件服务：确保 `backend/config.json` 使用生产可用的 SMTP。

## 常见问题
- 无法注册或收不到验证码：请检查 [backend/config.json](backend/config.json) 的 SMTP 配置是否正确，网络是否允许外发邮件。
- WebSocket 连接失败：请确认前端 `VITE_WEBSOCKET_URL` 使用 `ws://` 或 `wss://`，以及反代路径是否为 `/chess/ws`。
- CORS 报错（本地联调）：设置 `FRONTEND_URL=http://localhost:5173` 并重启后端；或统一走同源反代路径（推荐使用 Docker Compose 方式）。

## 开发参考
- 路由与中间件：见 [backend/route/route.go](backend/route/route.go) 与 [backend/middleware/auth.go](backend/middleware/auth.go)
- 数据库初始化与自动迁移：见 [backend/database/mysql.go](backend/database/mysql.go) 与 [backend/model/model.go](backend/model/model.go)
- WebSocket 接入：前端在应用启动后读取 `VITE_WEBSOCKET_URL` 并在登录成功后发起连接（见 [frontend/src/App.vue](frontend/src/App.vue) 与 [frontend/src/websocket/index.ts](frontend/src/websocket/index.ts)）

> 安全提示：仓库中的示例配置仅用于演示，切勿在公开仓库中提交真实账号、密码、密钥等敏感信息。如需对接生产环境，请妥善管理密钥并考虑完善 JWT 秘钥等安全机制。