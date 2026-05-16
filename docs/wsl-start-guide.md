# WSL 启动文档

本文档用于在 WSL 环境启动校园生活问答 MVP，包括 PostgreSQL + pgvector、embedding 服务、Go server、Web console 和 Wails 桌面客户端。

## 1. 环境要求

建议环境：

- WSL2 Ubuntu
- Docker Desktop 已开启 WSL integration，或 WSL 内已安装 Docker Engine
- Go 1.25+
- Node.js 18+
- npm
- make
- Wails CLI（仅开发桌面客户端时需要）

检查命令：

```bash
docker version
docker compose version
go version
node -v
npm -v
make -v
```

如果缺少基础工具，可在 WSL 中安装：

```bash
sudo apt update
sudo apt install -y make curl git
```

Node.js 建议使用 nvm 安装。Go 建议使用官方安装包或系统已配置版本。

如果要开发 Wails 桌面客户端，还需要安装 Linux GUI 依赖和 Wails CLI：

```bash
sudo apt install -y libgtk-3-dev libwebkit2gtk-4.0-dev gcc
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails version
```

如果提示找不到 `wails`，确认 Go bin 目录已加入 `PATH`：

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## 2. 启动基础服务

在项目根目录执行：

```bash
cd /home/kado_1/ans-b
docker compose -f deploy/docker-compose.yml up -d postgres embed-server
```

查看容器状态：

```bash
docker ps --filter name=campus-postgres --filter name=campus-embed-server
```

检查 embedding 服务：

```bash
curl -sS http://127.0.0.1:18080/healthz
```

首次启动 `embed-server` 会下载模型，时间会比较长。

## 3. 配置后端环境变量

复制配置模板：

```bash
cp .env.example .env
```

编辑 `.env`：

```bash
nano .env
```

默认配置如下：

```env
DATABASE_URL=postgres://campus:campus123@localhost:5432/campus_qa?sslmode=disable
EMBED_BASE_URL=http://127.0.0.1:18080
OPENAI_BASE_URL=https://api.kimi.com/coding/v1
OPENAI_MODEL=kimi-for-coding
OPENAI_API_KEY=replace-with-your-api-key
OPENAI_TIMEOUT_SECONDS=20
QA_MIN_SCORE=0.45
```

必须修改：

```env
OPENAI_API_KEY=你的 Kimi API Key
```

`.env` 已被 `.gitignore` 忽略，不会提交到仓库。

## 4. 导入问答种子数据

如果数据库还没有问答数据，执行：

```bash
cd /home/kado_1/ans-b/server
go run ./cmd/importqa -file ../data/qa_seed.json
```

确认数据：

```bash
docker exec campus-postgres psql -U campus -d campus_qa -c "SELECT COUNT(*) FROM knowledge_items; SELECT COUNT(*) FROM knowledge_chunks;"
```

## 5. 启动 Go Server

回到项目根目录：

```bash
cd /home/kado_1/ans-b
make server
```

`make server` 会先编译：

```text
server/bin/campus-server
```

然后启动 Go server。

默认监听：

```text
http://127.0.0.1:8080
```

检查：

```bash
curl -sS http://127.0.0.1:8080/healthz
```

测试问答接口：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/qa/ask \
  -H "Content-Type: application/json" \
  -d '{"question":"食堂几点关门？","limit":5}'
```

## 6. 启动 Web Console

新开一个 WSL 终端：

```bash
cd /home/kado_1/ans-b
make console
```

默认访问：

```text
http://127.0.0.1:5173/
```

console 默认 API 地址是：

```text
http://127.0.0.1:8080
```

## 7. 内网机器访问 Console

如果要让另一台机器访问 WSL 里的 console，需要监听所有网卡：

```bash
make console API_BASE_URL=http://你的WSL主机IP:8080
```

Makefile 默认 `HOST=0.0.0.0`，所以 console 会监听：

```text
http://你的WSL主机IP:5173/
```

示例：

```bash
make console API_BASE_URL=http://100.115.97.57:8080
```

另一台机器访问：

```text
http://100.115.97.57:5173/
```

如果使用项目里的 `lan` 命令：

```bash
make lan
```

它会使用 Makefile 中的：

```makefile
LAN_API_BASE_URL ?= http://100.115.97.57:8080
```

如果你的 WSL 主机 IP 变了，可以直接覆盖：

```bash
make lan LAN_API_BASE_URL=http://新的IP:8080
```

注意：

- 浏览器里的 `127.0.0.1` 永远指访问者自己的机器。
- 其他机器访问时，`API_BASE_URL` 必须写 Go server 所在机器的内网 IP。
- Go server 的 CORS 当前已允许 `http://100.*` 来源。

## 8. 开发 Wails 桌面客户端

桌面客户端目录：

```text
ans-b-client/
```

首次开发需要安装前端依赖：

```bash
cd /home/kado_1/ans-b/ans-b-client/frontend
npm install
```

开发模式启动：

```bash
cd /home/kado_1/ans-b/ans-b-client
wails dev
```

说明：

- `wails dev` 会自动启动客户端前端 Vite dev server。
- 桌面窗口会加载 `ans-b-client/frontend` 下的 Vue 页面。
- Wails 调试页面通常可在浏览器访问 `http://localhost:34115`。
- 如果客户端需要调用 Go server，需要先按第 5 节启动后端。

客户端前端单独开发：

```bash
cd /home/kado_1/ans-b/ans-b-client/frontend
npm run dev
```

生产构建：

```bash
cd /home/kado_1/ans-b/ans-b-client
wails build
```

构建产物：

```text
ans-b-client/build/bin/
```

客户端常见问题：

- `wails dev` 报 `libwebkit2gtk` 缺失：重新安装 `libwebkit2gtk-4.0-dev`。
- `wails` 命令不存在：确认 `$(go env GOPATH)/bin` 已加入 `PATH`。
- WSL 无法弹出桌面窗口：需要 Windows 11 WSLg，或本机已配置可用的 Linux GUI/X Server。

## 9. Console 和客户端区别

本项目当前有两个前端入口：

| 入口 | 路径 | 用途 | 启动命令 |
|---|---|---|---|
| Web console | `console/` | 管理和测试 MVP，包含知识录入、问答测试 | `make console` |
| Wails 客户端 | `ans-b-client/` | 桌面客户端开发入口 | `wails dev` |

目前 MVP 问答调试优先使用 `console/`。如果开发桌面应用体验，再进入 `ans-b-client/`。

## 10. 常用命令汇总

启动 Docker 服务：

```bash
docker compose -f deploy/docker-compose.yml up -d postgres embed-server
```

启动后端：

```bash
make server
```

启动本机 console：

```bash
make console
```

启动内网 console：

```bash
make console API_BASE_URL=http://你的WSL主机IP:8080
```

导入种子数据：

```bash
cd server
go run ./cmd/importqa -file ../data/qa_seed.json
```

运行测试：

```bash
make test
```

停止 Docker 服务：

```bash
docker compose -f deploy/docker-compose.yml down
```

启动 Wails 客户端：

```bash
cd ans-b-client
wails dev
```

构建 Wails 客户端：

```bash
cd ans-b-client
wails build
```

## 11. 常见问题

### 11.1 其他机器打不开 `5173`

确认 console 是否用 `0.0.0.0` 启动：

```bash
make console API_BASE_URL=http://你的WSL主机IP:8080
```

确认防火墙或网络是否允许访问 `5173`。

### 11.2 页面打开了，但请求 API 失败

检查页面顶部显示的 API 地址。其他机器访问时不能是：

```text
http://127.0.0.1:8080
```

应该是：

```text
http://你的WSL主机IP:8080
```

### 11.3 提问返回没有 AI 回答

检查 `.env` 是否配置：

```env
OPENAI_API_KEY=你的 Kimi API Key
OPENAI_BASE_URL=https://api.kimi.com/coding/v1
OPENAI_MODEL=kimi-for-coding
```

修改 `.env` 后需要重启 Go server。

### 11.4 随便输入也有候选结果

这是向量检索的正常现象。系统会展示 Top 候选，但只有最高相似度达到 `QA_MIN_SCORE` 才会正式回答。

默认：

```env
QA_MIN_SCORE=0.45
```
