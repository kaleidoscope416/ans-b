# 校园生活百事通数据库设计文档

## 1. 文档目的

本文档描述校园生活百事通系统第一版数据库设计，包括表结构、字段、关系、索引和状态枚举。本文档只描述数据库设计，不描述接口细节和业务流程。

## 2. 数据库选型

数据库使用 PostgreSQL，并启用 pgvector 扩展。

用途：

1. 保存学生用户、知识库、投稿、日志和管理员数据。
2. 保存知识文本向量。
3. 支持关键词检索和语义相似度检索。

## 3. 表关系概览

```text
admin_users

users

knowledge_items 1 ─── N knowledge_chunks

knowledge_items 1 ─── N query_logs

users 1 ─── N user_submissions

users 1 ─── N query_logs

user_submissions 审核通过后生成 knowledge_items
```

## 4. 表结构设计

### 4.1 管理员表 `admin_users`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | bigserial | primary key | 管理员 ID |
| username | varchar(64) | unique, not null | 管理员账号 |
| password_hash | varchar(255) | not null | 密码哈希 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

### 4.2 学生用户表 `users`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | bigserial | primary key | 学生用户 ID |
| username | varchar(64) | unique, not null | 学生账号 |
| password_hash | varchar(255) | not null | 密码哈希 |
| nickname | varchar(100) |  | 昵称 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

### 4.3 知识主表 `knowledge_items`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | bigserial | primary key | 知识 ID |
| title | varchar(200) | not null | 标题 |
| question | text | not null | 标准问题 |
| answer | text | not null | 标准答案 |
| category | varchar(100) |  | 分类 |
| tags | text[] |  | 标签 |
| source_type | varchar(32) | not null | 来源类型 |
| status | varchar(32) | not null | 状态 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

### 4.4 知识片段表 `knowledge_chunks`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | bigserial | primary key | 片段 ID |
| item_id | bigint | foreign key | 关联知识 ID |
| chunk_text | text | not null | 片段文本 |
| embedding | vector |  | 文本向量 |
| source_url | text |  | 来源地址 |
| page_no | integer |  | 页码 |
| created_at | timestamptz | not null | 创建时间 |

说明：`embedding` 维度由实际 embedding 模型决定，例如 `vector(1536)` 或 `vector(1024)`。

### 4.5 用户投稿表 `user_submissions`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | bigserial | primary key | 投稿 ID |
| user_id | bigint | foreign key, not null | 投稿学生用户 ID |
| question | text | not null | 投稿问题 |
| answer | text | not null | 参考答案 |
| category | varchar(100) |  | 分类 |
| tags | text[] |  | 标签 |
| source | text |  | 信息来源 |
| remark | text |  | 备注 |
| status | varchar(32) | not null | 投稿状态 |
| reviewer_note | text |  | 审核备注 |
| created_at | timestamptz | not null | 创建时间 |
| reviewed_at | timestamptz |  | 审核时间 |

### 4.6 查询日志表 `query_logs`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | bigserial | primary key | 日志 ID |
| user_id | bigint | foreign key | 学生用户 ID，匿名查询时为空 |
| user_question | text | not null | 用户原始问题 |
| normalized_question | text |  | 归一化问题 |
| intent | varchar(64) |  | 问题意图 |
| matched_item_id | bigint | foreign key | 命中的知识 ID |
| hit_score | numeric(5,4) |  | 命中分数 |
| created_at | timestamptz | not null | 创建时间 |

## 5. 状态枚举

### 5.1 知识来源 `source_type`

| 值 | 说明 |
|---|---|
| faq | 管理员手动录入或 FAQ 导入 |
| document | 文档解析生成 |
| user_submit | 用户投稿审核生成 |

### 5.2 知识状态 `knowledge_items.status`

| 值 | 说明 |
|---|---|
| draft | 草稿 |
| pending | 待审核 |
| approved | 已通过，可被检索 |
| disabled | 已禁用，不可被检索 |

### 5.3 投稿状态 `user_submissions.status`

| 值 | 说明 |
|---|---|
| pending | 待审核 |
| approved | 已通过 |
| rejected | 已驳回 |

## 6. 索引设计

### 6.1 管理员表

```sql
CREATE UNIQUE INDEX idx_admin_users_username ON admin_users(username);
```

### 6.2 学生用户表

```sql
CREATE UNIQUE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 6.3 知识主表

```sql
CREATE INDEX idx_knowledge_items_status ON knowledge_items(status);
CREATE INDEX idx_knowledge_items_category ON knowledge_items(category);
CREATE INDEX idx_knowledge_items_source_type ON knowledge_items(source_type);
CREATE INDEX idx_knowledge_items_created_at ON knowledge_items(created_at);
```

### 6.4 知识片段表

```sql
CREATE INDEX idx_knowledge_chunks_item_id ON knowledge_chunks(item_id);
```

向量索引示例：

```sql
CREATE INDEX idx_knowledge_chunks_embedding
ON knowledge_chunks
USING ivfflat (embedding vector_cosine_ops);
```

### 6.5 用户投稿表

```sql
CREATE INDEX idx_user_submissions_user_id ON user_submissions(user_id);
CREATE INDEX idx_user_submissions_status ON user_submissions(status);
CREATE INDEX idx_user_submissions_created_at ON user_submissions(created_at);
```

### 6.6 查询日志表

```sql
CREATE INDEX idx_query_logs_user_id ON query_logs(user_id);
CREATE INDEX idx_query_logs_created_at ON query_logs(created_at);
CREATE INDEX idx_query_logs_matched_item_id ON query_logs(matched_item_id);
CREATE INDEX idx_query_logs_intent ON query_logs(intent);
```

## 7. 数据约束

1. 管理员账号必须唯一。
2. 学生账号必须唯一。
3. 学生和管理员密码必须保存哈希值。
4. 知识标题、问题和答案不能为空。
5. 投稿必须绑定学生用户。
6. 投稿问题和答案不能为空。
7. 只有 `approved` 状态的知识可以进入学生问答检索范围。
8. 删除知识时应同步处理关联知识片段。
9. 查询日志写入失败不应影响问答结果返回。

## 8. 初始化要求

数据库初始化应完成：

1. 创建 pgvector 扩展。
2. 创建全部业务表。
3. 创建必要索引。
4. 初始化至少一个管理员账号。
5. 可选初始化演示学生账号。
6. 可选导入演示 FAQ 数据。

初始化扩展示例：

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```
