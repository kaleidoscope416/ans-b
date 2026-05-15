# ans-b-client 启动指南（Linux）

本文档面向团队成员，介绍如何在 **Linux** 环境下配置开发环境并启动 Wails 桌面客户端。

---

## 项目技术栈

| 层级 | 技术 |
|------|------|
| 桌面端框架 | [Wails v2](https://wails.io/)（Go + Webview） |
| 前端框架 | Vue 3 + Vite |
| 前端包管理 | npm |
| Go 版本 | ≥ 1.23.0 |

> 后端服务（`../server/`）使用 Gin 框架，Go 版本 ≥ 1.25.0，可独立启动。

---

## 一、前置依赖安装

### 1.1 系统级依赖（必需）

Wails v2 在 Linux 上编译需要 GTK3 与 WebKit2GTK 开发库。请根据你的发行版执行对应命令：

**Debian / Ubuntu**
```bash
sudo apt update
sudo apt install -y libgtk-3-dev libwebkit2gtk-4.0-dev gcc
```

**Arch / Manjaro**
```bash
sudo pacman -S gtk3 webkit2gtk gcc
```

**Fedora**
```bash
sudo dnf install gtk3-devel webkit2gtk3-devel gcc
```

### 1.2 Go

安装 Go **1.23.0 或更高版本**：

```bash
# 下载（以 1.23.4 为例，建议去官网查看最新稳定版）
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz

# 解压到 /usr/local
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz

# 配置环境变量（写入 ~/.bashrc 或 ~/.zshrc）
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# 验证
source ~/.bashrc   # 或 source ~/.zshrc
go version
```

> 如已安装 Go，请确保版本 `≥ 1.23.0`。

### 1.3 Node.js & npm

前端构建需要 Node.js 与 npm，推荐安装 LTS 版本：

```bash
# 使用 nvm 安装（推荐）
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc
nvm install --lts
nvm use --lts

# 验证
node -v
npm -v
```

### 1.4 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 验证
wails version
```

> 安装后若找不到 `wails` 命令，请确认 `$GOPATH/bin` 已加入 `PATH`。

---

## 二、项目初始化

```bash
# 1. 进入客户端目录
cd ans-b-client

# 2. 安装前端依赖（首次启动必须）
cd frontend && npm install && cd ..

# 3. 下载 Go 依赖
wails dev        # 开发模式会自动拉取依赖
# 或手动执行
go mod tidy
```

---

## 三、启动项目

### 开发模式（推荐日常开发）

```bash
cd ans-b-client
wails dev
```

- 会自动启动 Vite 开发服务器，支持前端热重载。
- Go 代码变更后也会自动重新编译。
- 浏览器可访问 http://localhost:34115 调试前端，并调用 Go 方法。

### 生产构建

```bash
cd ans-b-client
wails build
```

构建产物位于 `build/bin/` 目录下，可直接运行：

```bash
./build/bin/ans-b-client
```

---

## 四、后端服务启动（可选）

如果本地需要同时运行后端 API：

```bash
cd server
go mod tidy
go run main.go
```

> 后端默认端口请查看 `server/main.go` 中的配置，客户端可通过配置项指向本地后端地址。

---

## 五、常见问题

### Q1: `wails dev` 报错找不到 `libwebkit2gtk`

确认已安装对应发行版的 `libwebkit2gtk-4.0-dev`（或 `webkit2gtk`）包，并重启终端后再试。

### Q2: `go install wails` 下载缓慢

设置 Go 国内代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### Q3: 前端 `npm install` 失败

尝试切换 npm 镜像源：

```bash
npm config set registry https://registry.npmmirror.com
```

### Q4: 构建时提示 `gcc` 命令未找到

安装系统编译工具链：

```bash
# Debian/Ubuntu
sudo apt install build-essential

# Arch
sudo pacman -S base-devel

# Fedora
sudo dnf groupinstall "Development Tools"
```

---

## 六、目录结构速览

```
ans-b-client/
├── app.go              # Go 后端逻辑（绑定给前端调用的方法）
├── main.go             # Wails 程序入口
├── wails.json          # Wails 项目配置
├── go.mod / go.sum     # Go 依赖
├── frontend/
│   ├── package.json    # 前端依赖
│   ├── vite.config.js  # Vite 配置
│   └── src/            # Vue 源码
└── build/              # 构建输出
```

---

## 参考链接

- [Wails 官方文档](https://wails.io/docs/)
- [Vue 3 文档](https://vuejs.org/)
- [Go 安装指南](https://go.dev/doc/install)
