# 校园生活百事通

这是一个 RAG 架构的校园智能问答助手，由大一天文社社员星见 遥（Hoshimi Haruka）给你解答疑惑。

项目面向校园生活场景，支持将校园常见问答导入知识库，通过向量检索找到最相关的知识，再结合 OpenAI 兼容大模型生成更自然的回答。系统当前覆盖知识录入、向量化入库、问题检索、候选结果展示和 AI 增强回答。

## 架构

```text
┌────────────────────────────────────────────────────────────┐
│                        客户端层                              │
│  Web Console                       Wails 桌面客户端          │
│  知识录入 / 问答测试                 学生侧桌面应用            │
└───────────────┬──────────────────────────────┬─────────────┘
                │ HTTP / JSON                  │ HTTP / JSON
                ▼                              ▼
┌────────────────────────────────────────────────────────────┐
│                      Go Server 服务层                        │
│  Gin API                                                    │
│  ├─ 知识录入：校验、组装 chunk、向量化、入库                  │
│  ├─ 问答检索：问题向量化、pgvector TopN、相似度阈值判断        │
│  └─ AI 增强：把候选知识和相似度交给 OpenAI 兼容 API           │
└───────────────┬──────────────────────────────┬─────────────┘
                │                              │
                │ Embedding HTTP               │ SQL / Vector Query
                ▼                              ▼
┌──────────────────────────────┐    ┌─────────────────────────┐
│       Embedding 服务          │    │ PostgreSQL + pgvector    │
│ FastAPI + bge-large-zh-v1.5  │    │ 业务数据 / 知识库 / 向量   │
└──────────────────────────────┘    └─────────────────────────┘
                │
                │ 检索结果 + 相似度
                ▼
┌────────────────────────────────────────────────────────────┐
│                  OpenAI-compatible API                      │
│       Kimi / OpenAI / 其他兼容 Chat Completions 服务          │
└────────────────────────────────────────────────────────────┘
```

核心流程：

- 管理端录入知识：问题、答案、分类、标签。
- 后端调用 embedding 服务，把知识内容转换为向量。
- 向量写入 PostgreSQL 的 `pgvector` 字段。
- 用户提问时，后端先向量化问题，再用 pgvector 做相似度检索。
- 相似度达到阈值后，将候选知识和相似度交给大模型生成回答。
- 相似度不足时，不强行回答，只展示候选并提示无法确认。

## 技术栈

| 模块 | 技术 |
|---|---|
| Web Console | Vue 3 + Vite + TDesign Vue Next |
| 桌面客户端 | Wails v2 + Vue 3 |
| 后端服务 | Go + Gin |
| 数据库 | PostgreSQL |
| 登录状态 | Redis |
| 向量检索 | pgvector |
| Embedding 服务 | FastAPI + sentence-transformers |
| Embedding 模型 | `BAAI/bge-large-zh-v1.5` |
| AI 回答 | OpenAI-compatible Chat Completions API |
| 默认 AI Provider | Kimi API |
| 容器化 | Docker Compose |

## 目录结构

```text
.
├── console/          # Web Console，知识录入和问答测试
├── server/           # Go 后端服务
├── deploy/           # PostgreSQL、pgvector、embedding 服务 Docker 配置
├── data/             # 种子问答数据
├── docs/             # 项目文档
└── ans-b-client/     # Wails 桌面客户端
```

## 快速启动

WSL 环境请参考：

[docs/wsl-start-guide.md](docs/wsl-start-guide.md)

常用命令：

```bash
docker compose -f deploy/docker-compose.yml up -d postgres redis embed-server
cp .env.example .env
make server
make console
```

`.env` 中需要配置 Kimi API Key：

```env
OPENAI_BASE_URL=https://api.kimi.com/coding/v1
OPENAI_MODEL=kimi-for-coding
OPENAI_API_KEY=replace-with-your-api-key
```

后端登录状态存储在 Redis 中，默认连接 `localhost:6379`，可通过 `.env` 中的 `REDIS_ADDR`、`REDIS_PASSWORD`、`REDIS_DB` 调整。
