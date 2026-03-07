# pjsk-bot

## 使用 Docker 启动

通过 `docker-compose.yml` 启动 4 个容器：

- `pjsk-elasticsearch` 保存歌曲消息
- `pjsk-server`        歌曲查询后端服务
- `pjsk-nonebot`       机器人，AI后台
- `pjsk-napcat`        NapCat QQ客户端

并且所有服务都挂在同一个自定义网络：`pjsk-net`。

## 配置说明

Python 侧请求 `server` 的地址已统一提取到 `pjsk-bot/config.py`，默认从 `pjsk-bot/application-docker.yaml` 读取。

默认配置文件内容：

```yaml
server:
  base_url: "http://server:9470"
```

- Docker 场景下建议保持以上值（通过服务名 `server` 访问）
- 如果要切换配置文件，可设置环境变量：`PJSK_CONFIG_FILE=/app/pjsk-bot/application-docker.yaml`

## 前置条件

- 已安装 Docker 与 Docker Compose（`docker compose`）
- 项目根目录存在 `.env.prod`，至少包含：

```env
DRIVER=~fastapi+~websockets
HOST=0.0.0.0
PORT=9350
```

- 在`docker-compose.yml`的第44行把`- OPENAI_API_KEY=${YOUR_OPENAI_API_KEY}`换成自己的API Key
- 连接地址，模型名称可以在`pjsk-bot/application-docker.yaml`配置
- 数据服务器相关的配置可以在`server/resources/config/application.yaml`配置

## 构建并启动

在项目根目录执行：

```bash
docker compose up -d --build
```

## 检查服务与网络

查看服务状态：

```bash
docker compose ps
```

查看网络：

```bash
docker network inspect pjsk-net
```

## 查看日志

```bash
docker compose logs -f nonebot
docker compose logs -f napcat
docker compose logs -f server
docker compose logs -f elasticsearch
```

## NapCat 反向 WebSocket 配置

在 NacCat 容器的启动日志可以找到带有token的 WebUI 的url

需要在 NapCat 配置反向 WebSocket客户端

由于容器在同一网络中，NapCat 可直接通过服务名访问 NoneBot：

```text
ws://nonebot:9350/onebot/v11/ws
```

不要使用 `127.0.0.1`，它只指向容器自身。

## 停止与清理

停止：

```bash
docker compose down
```

停止并删除 ES 数据卷：

```bash
docker compose down -v
```

## 查歌说明

第一次启动项目，查找歌曲时没有数据会返回错误

可以给`/pjsk/update`发一个POST请求更新数据

或者直接让AI调用更新歌曲数据库的工具