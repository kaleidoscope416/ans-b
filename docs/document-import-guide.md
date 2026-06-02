# 文档 JSON 导入指南

本文说明如何把 `data/scut_pages.json` 中的官网文档导入到数据库。

## 1. 前置服务

先启动 PostgreSQL 和本地向量服务：

```bash
docker ps
```

需要看到以下容器：

```text
campus-postgres
campus-embed-server
```

如果没有启动，先在项目根目录执行：

```bash
cd /home/kado_1/ans-b
docker compose -f deploy/docker-compose.yml up -d
```

## 2. JSON 文件格式

导入文件路径：

```text
/home/kado_1/ans-b/data/scut_pages.json
```

格式是数组，每条文档包含：

```json
{
  "title": "华南理工大学VPN系统使用说明",
  "content": "正文内容……",
  "category": "信息化服务",
  "source_name": "华南理工大学网络与信息化办公室",
  "source_url": "https://web.scut.edu.cn/2020/1123/c15298a409464/page.htm",
  "tags": ["VPN", "校园网", "校外访问"]
}
```

必填字段：

```text
title
content
source_url
```

## 3. 导入命令

如果 `.env` 已配置：

```env
DATABASE_URL=postgres://campus:campus123@localhost:5432/campus_qa?sslmode=disable
EMBED_BASE_URL=http://127.0.0.1:18080
```

执行：

```bash
cd /home/kado_1/ans-b/server
go run ./cmd/importdocs -file ../data/scut_pages.json
```

也可以显式指定数据库和向量服务：

```bash
cd /home/kado_1/ans-b/server
go run ./cmd/importdocs \
  -file ../data/scut_pages.json \
  -db "postgres://campus:campus123@localhost:5432/campus_qa?sslmode=disable" \
  -embed-url "http://127.0.0.1:18080" \
  -page-batch-size 50 \
  -embed-batch-size 2 \
  -chunk-batch-size 500
```

## 4. 入库行为

导入流程：

```text
读取 scut_pages.json
-> 校验 title/content/source_url
-> 每篇文档按 300 字切片，60 字重叠
-> 按 page-batch-size 分批处理文档
-> 按 embed-batch-size 调用 embed 服务生成向量
-> 一个批次内写入或更新 knowledge_items
-> 删除本批文档旧 chunks
-> 按 chunk-batch-size 多行 INSERT 写入 knowledge_chunks
```

字段映射：

```text
knowledge_items.title = title
knowledge_items.question = title
knowledge_items.answer = content
knowledge_items.category = category
knowledge_items.tags = tags
knowledge_items.source_type = official_page
knowledge_chunks.chunk_text = 正文切片
knowledge_chunks.source_url = source_url
```

同一个 `source_url` 重复导入时，会更新旧文档并重建 chunks，不会不断新增重复数据。
