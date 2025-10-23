# 中国象棋对弈网站

克隆此仓库

在`backend`服务中，修改环境变量`FRONTEND_URL`为你的前端地址。

然后在`chinese-chess-frontend`文件夹中创建`.env`文件，内容如下：

```
VITE_API_URL=http://localhost:8080/api
VITE_WEBSOCKET_URL=http://localhost:8080/ws
```

请改为实际的地址。

完成修改后，执行

```sh
docker compose up -d
```

如果需要使用https或者域名访问，建议使用反向代理