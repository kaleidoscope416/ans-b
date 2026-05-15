# 校园生活百事通接口文档

## 1. 文档目的

本文档描述校园生活百事通系统第一版 REST API，包括接口分组、请求参数、响应参数和错误码。本文档只描述接口，不描述数据库表结构和内部架构。

## 2. 接口汇总

### 2.1 公共接口

公共接口不要求登录。

| 方法 | 路径 | 功能 | 鉴权 |
|---|---|---|---|
| POST | `/api/qa/ask` | 提交自然语言问题并获取答案 | 否 |
| GET | `/api/hot-questions` | 获取热点问题排行榜 | 否 |
| POST | `/api/users/register` | 学生注册 | 否 |
| POST | `/api/users/login` | 学生登录 | 否 |

### 2.2 学生接口

学生接口供学生客户端调用，需要学生 JWT 鉴权。

| 方法 | 路径 | 功能 | 鉴权 |
|---|---|---|---|
| GET | `/api/users/me` | 获取当前学生信息 | 是 |
| POST | `/api/submissions` | 提交新的问答内容 | 是 |
| GET | `/api/users/me/submissions` | 获取自己的投稿历史 | 是 |

### 2.3 管理员接口

管理员接口供 `admin-console` 调用，用于知识库维护和投稿审核。除登录接口外都需要 JWT 鉴权。

| 方法 | 路径 | 功能 | 鉴权 |
|---|---|---|---|
| POST | `/api/admin/login` | 管理员登录 | 否 |
| GET | `/api/admin/knowledge` | 查询知识列表 | 是 |
| POST | `/api/admin/knowledge` | 新增知识 | 是 |
| PUT | `/api/admin/knowledge/{id}` | 编辑知识 | 是 |
| DELETE | `/api/admin/knowledge/{id}` | 删除知识 | 是 |
| POST | `/api/admin/knowledge/import` | 批量导入 FAQ | 是 |
| GET | `/api/admin/submissions` | 查询投稿列表 | 是 |
| POST | `/api/admin/submissions/{id}/approve` | 审核通过投稿 | 是 |
| POST | `/api/admin/submissions/{id}/reject` | 驳回投稿 | 是 |

### 2.4 接口数量统计

| 类型 | 数量 |
|---|---:|
| 公共接口 | 4 |
| 学生接口 | 3 |
| 管理员接口 | 9 |
| 合计 | 16 |

## 3. 通用约定

### 3.1 基础路径

```text
/api
```

### 3.2 请求格式

除文件上传接口外，请求体统一使用 JSON。

```http
Content-Type: application/json
```

### 3.3 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 3.4 鉴权方式

学生客户端调用学生接口时需要携带学生 JWT，`admin-console` 调用管理员接口时需要携带管理员 JWT。

```http
Authorization: Bearer <token>
```

## 4. 错误码

| 错误码 | 说明 |
|---|---|
| 0 | 成功 |
| 40000 | 请求参数错误 |
| 40001 | 未登录或登录已过期 |
| 40003 | 无权限访问 |
| 40400 | 资源不存在 |
| 40900 | 数据冲突 |
| 50000 | 服务内部错误 |
| 50010 | 检索服务不可用 |
| 50020 | 模型服务不可用 |

## 5. 问答接口

### 5.1 提交问题

```text
POST /api/qa/ask
```

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| question | string | 是 | 用户问题 |

请求示例：

```json
{
  "question": "食堂几点关门？"
}
```

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| answer | string | 回答内容 |
| matched_question | string | 匹配到的知识问题 |
| score | number | 匹配分数 |
| intent | string | 识别出的意图 |
| related | array | 相关问题列表 |
| hit | boolean | 是否命中高置信度答案 |

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "answer": "一食堂晚餐营业至 20:00，二食堂营业至 21:00。",
    "matched_question": "食堂营业时间是什么？",
    "score": 0.86,
    "intent": "查询时间",
    "related": [
      "一食堂在哪里？",
      "校园卡可以在食堂充值吗？"
    ],
    "hit": true
  }
}
```

## 6. 热点问题接口

### 6.1 获取热点问题

```text
GET /api/hot-questions
```

查询参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| limit | number | 否 | 返回数量，默认 10 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| items | array | 热点问题列表 |

热点问题项：

| 字段 | 类型 | 说明 |
|---|---|---|
| question | string | 问题 |
| count | number | 查询次数 |
| category | string | 分类 |

## 7. 学生账户接口

### 7.1 学生注册

```text
POST /api/users/register
```

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| username | string | 是 | 学生账号 |
| password | string | 是 | 学生密码 |
| nickname | string | 否 | 昵称 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 学生用户 ID |
| username | string | 学生账号 |
| nickname | string | 昵称 |

### 7.2 学生登录

```text
POST /api/users/login
```

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| username | string | 是 | 学生账号 |
| password | string | 是 | 学生密码 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| token | string | 学生 JWT |
| expires_in | number | 过期秒数 |

### 7.3 获取当前学生信息

```text
GET /api/users/me
```

鉴权：需要学生 JWT。

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 学生用户 ID |
| username | string | 学生账号 |
| nickname | string | 昵称 |

## 8. 用户投稿接口

### 8.1 提交新问答

```text
POST /api/submissions
```

鉴权：需要学生 JWT。

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| question | string | 是 | 投稿问题 |
| answer | string | 是 | 参考答案 |
| category | string | 否 | 分类 |
| tags | array | 否 | 标签 |
| source | string | 否 | 信息来源 |
| remark | string | 否 | 备注 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 投稿 ID |
| status | string | 投稿状态 |

### 8.2 获取自己的投稿历史

```text
GET /api/users/me/submissions
```

鉴权：需要学生 JWT。

查询参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| status | string | 否 | 投稿状态 |
| page | number | 否 | 页码 |
| page_size | number | 否 | 每页数量 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| items | array | 投稿记录列表 |
| total | number | 总数量 |

投稿记录项：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 投稿 ID |
| question | string | 投稿问题 |
| answer | string | 参考答案 |
| status | string | 投稿状态 |
| reviewer_note | string | 审核备注 |
| created_at | string | 提交时间 |
| reviewed_at | string | 审核时间 |

## 9. 管理员鉴权接口

### 9.1 管理员登录

```text
POST /api/admin/login
```

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| username | string | 是 | 管理员账号 |
| password | string | 是 | 管理员密码 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| token | string | JWT |
| expires_in | number | 过期秒数 |

## 10. 知识库管理接口

### 10.1 查询知识列表

```text
GET /api/admin/knowledge
```

查询参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| keyword | string | 否 | 搜索关键词 |
| category | string | 否 | 分类 |
| status | string | 否 | 状态 |
| page | number | 否 | 页码 |
| page_size | number | 否 | 每页数量 |

### 10.2 新增知识

```text
POST /api/admin/knowledge
```

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| title | string | 是 | 标题 |
| question | string | 是 | 问题 |
| answer | string | 是 | 答案 |
| category | string | 否 | 分类 |
| tags | array | 否 | 标签 |
| status | string | 是 | 状态 |

### 10.3 编辑知识

```text
PUT /api/admin/knowledge/{id}
```

路径参数：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 知识 ID |

请求参数同新增知识。

### 10.4 删除知识

```text
DELETE /api/admin/knowledge/{id}
```

路径参数：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 知识 ID |

### 10.5 批量导入 FAQ

```text
POST /api/admin/knowledge/import
```

请求格式：

```http
Content-Type: multipart/form-data
```

表单参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| file | file | 是 | CSV 或 Excel 文件 |

响应数据：

| 字段 | 类型 | 说明 |
|---|---|---|
| success_count | number | 成功数量 |
| failed_count | number | 失败数量 |
| errors | array | 失败原因列表 |

## 11. 投稿审核接口

### 11.1 查询投稿列表

```text
GET /api/admin/submissions
```

查询参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| status | string | 否 | 投稿状态 |
| page | number | 否 | 页码 |
| page_size | number | 否 | 每页数量 |

### 11.2 审核通过投稿

```text
POST /api/admin/submissions/{id}/approve
```

路径参数：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 投稿 ID |

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| reviewer_note | string | 否 | 审核备注 |

### 11.3 驳回投稿

```text
POST /api/admin/submissions/{id}/reject
```

路径参数：

| 字段 | 类型 | 说明 |
|---|---|---|
| id | number | 投稿 ID |

请求参数：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| reviewer_note | string | 否 | 驳回原因 |
