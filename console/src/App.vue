<script setup>
import { computed, reactive, ref } from 'vue'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:23456'

const knowledgeForm = reactive({
  question: '',
  answer: '',
  category: '',
  tags: '',
  source: '',
  remark: '',
})

const askForm = reactive({
  question: '',
})

const adminForm = reactive({
  username: '',
  password: '',
})

const reviewForm = reactive({
  question: '',
  answer: '',
  category: '',
  tags: '',
  source: '',
  remark: '',
  reviewerNote: '',
})

const submitLoading = ref(false)
const askLoading = ref(false)
const adminLoading = ref(false)
const submissionsLoading = ref(false)
const reviewLoading = ref(false)
const submitMessage = ref('')
const submitError = ref('')
const askError = ref('')
const askResult = ref(null)
const adminMessage = ref('')
const adminError = ref('')
const reviewMessage = ref('')
const reviewError = ref('')
const adminToken = ref(window.localStorage.getItem('ans-b-admin-token') || '')
const adminUser = ref(JSON.parse(window.localStorage.getItem('ans-b-admin-user') || 'null'))
const reviewStatus = ref('pending')
const submissions = ref([])
const selectedSubmission = ref(null)

const canSubmitKnowledge = computed(() => (
  knowledgeForm.question.trim() &&
  knowledgeForm.answer.trim() &&
  !submitLoading.value
))

const canAsk = computed(() => askForm.question.trim() && !askLoading.value)
const isAdminLoggedIn = computed(() => Boolean(adminToken.value))
const canLoginAdmin = computed(() => (
  adminForm.username.trim() &&
  adminForm.password.trim() &&
  !adminLoading.value
))
const canApproveSubmission = computed(() => (
  selectedSubmission.value &&
  reviewForm.question.trim() &&
  reviewForm.answer.trim() &&
  !reviewLoading.value
))
const canRejectSubmission = computed(() => selectedSubmission.value && !reviewLoading.value)

function parseTags(value) {
  return value
    .split(/[,，\n]/)
    .map((tag) => tag.trim())
    .filter(Boolean)
}

async function requestJSON(path, options = {}) {
  const {
    method = 'GET',
    payload,
    auth = false,
  } = options
  const headers = {}
  if (payload !== undefined) {
    headers['Content-Type'] = 'application/json'
  }
  if (auth && adminToken.value) {
    headers.Authorization = `Bearer ${adminToken.value}`
  }
  const response = await fetch(`${apiBaseURL}${path}`, {
    method,
    headers,
    body: payload === undefined ? undefined : JSON.stringify(payload),
  })
  const result = await response.json().catch(() => null)
  if (!response.ok || result?.code !== 0) {
    throw new Error(result?.message || `HTTP ${response.status}`)
  }
  return result.data
}

async function postJSON(path, payload) {
  return requestJSON(path, { method: 'POST', payload })
}

async function loginAdmin() {
  if (!canLoginAdmin.value) return

  adminLoading.value = true
  adminMessage.value = ''
  adminError.value = ''

  try {
    const result = await requestJSON('/api/v1/auth/admin/login', {
      method: 'POST',
      payload: {
        username: adminForm.username.trim(),
        password: adminForm.password.trim(),
      },
    })
    adminToken.value = result.token
    adminUser.value = result.user
    window.localStorage.setItem('ans-b-admin-token', result.token)
    window.localStorage.setItem('ans-b-admin-user', JSON.stringify(result.user))
    adminForm.password = ''
    adminMessage.value = `已登录：${result.user?.username || adminForm.username.trim()}`
    await loadSubmissions()
  } catch (error) {
    adminError.value = error.message
  } finally {
    adminLoading.value = false
  }
}

function logoutAdmin() {
  adminToken.value = ''
  adminUser.value = null
  submissions.value = []
  selectedSubmission.value = null
  window.localStorage.removeItem('ans-b-admin-token')
  window.localStorage.removeItem('ans-b-admin-user')
}

function fillReviewForm(submission) {
  selectedSubmission.value = submission
  reviewForm.question = submission?.question || ''
  reviewForm.answer = submission?.answer || ''
  reviewForm.category = submission?.category || ''
  reviewForm.tags = Array.isArray(submission?.tags) ? submission.tags.join('，') : ''
  reviewForm.source = submission?.source || ''
  reviewForm.remark = submission?.remark || ''
  reviewForm.reviewerNote = submission?.reviewer_note || ''
  reviewMessage.value = ''
  reviewError.value = ''
}

async function loadSubmissions() {
  if (!isAdminLoggedIn.value) return

  submissionsLoading.value = true
  reviewError.value = ''

  try {
    const query = reviewStatus.value ? `?status=${encodeURIComponent(reviewStatus.value)}` : ''
    submissions.value = await requestJSON(`/api/v1/submissions${query}`, { auth: true })
    if (!submissions.value.some((item) => item.id === selectedSubmission.value?.id)) {
      fillReviewForm(submissions.value[0] || null)
    }
  } catch (error) {
    reviewError.value = error.message
  } finally {
    submissionsLoading.value = false
  }
}

async function approveSubmission() {
  if (!canApproveSubmission.value) return

  reviewLoading.value = true
  reviewMessage.value = ''
  reviewError.value = ''

  try {
    await requestJSON(`/api/v1/submissions/${selectedSubmission.value.id}/approve`, {
      method: 'POST',
      auth: true,
      payload: {
        question: reviewForm.question.trim(),
        answer: reviewForm.answer.trim(),
        category: reviewForm.category.trim(),
        tags: parseTags(reviewForm.tags),
        source: reviewForm.source.trim(),
        remark: reviewForm.remark.trim(),
        reviewer_note: reviewForm.reviewerNote.trim(),
      },
    })
    reviewMessage.value = '审核通过，已生成向量并进入知识库'
    await loadSubmissions()
  } catch (error) {
    reviewError.value = error.message
  } finally {
    reviewLoading.value = false
  }
}

async function rejectSubmission() {
  if (!canRejectSubmission.value) return

  reviewLoading.value = true
  reviewMessage.value = ''
  reviewError.value = ''

  try {
    await requestJSON(`/api/v1/submissions/${selectedSubmission.value.id}/reject`, {
      method: 'POST',
      auth: true,
      payload: {
        reviewer_note: reviewForm.reviewerNote.trim(),
      },
    })
    reviewMessage.value = '已驳回投稿'
    await loadSubmissions()
  } catch (error) {
    reviewError.value = error.message
  } finally {
    reviewLoading.value = false
  }
}

async function submitKnowledge() {
  if (!canSubmitKnowledge.value) return

  submitLoading.value = true
  submitMessage.value = ''
  submitError.value = ''

  try {
    await postJSON('/api/v1/knowledge', {
      question: knowledgeForm.question.trim(),
      answer: knowledgeForm.answer.trim(),
      category: knowledgeForm.category.trim(),
      tags: parseTags(knowledgeForm.tags),
      source: knowledgeForm.source.trim(),
      remark: knowledgeForm.remark.trim(),
    })
    submitMessage.value = '知识已写入数据库并完成向量化'
    knowledgeForm.question = ''
    knowledgeForm.answer = ''
    knowledgeForm.tags = ''
    knowledgeForm.remark = ''
  } catch (error) {
    submitError.value = error.message
  } finally {
    submitLoading.value = false
  }
}

async function askQuestion() {
  if (!canAsk.value) return

  askLoading.value = true
  askError.value = ''
  askResult.value = null

  try {
    askResult.value = await postJSON('/api/v1/qa/ask', {
      question: askForm.question.trim(),
      limit: 5,
    })
  } catch (error) {
    askError.value = error.message
  } finally {
    askLoading.value = false
  }
}

function candidateTitle(item) {
  return item?.title || item?.matched_question || `知识 #${item?.item_id || item?.id || ''}`
}

function candidateBody(item) {
  return item?.chunk_text || item?.answer || ''
}
</script>

<template>
  <main class="console-page">
    <header class="console-header">
      <div>
        <h1>校园生活百事通 Console</h1>
        <p>审核用户投稿，通过后进入知识库并生成向量。</p>
      </div>
      <t-tag theme="primary" variant="light">API {{ apiBaseURL }}</t-tag>
    </header>

    <section class="panel review-panel">
      <div class="panel-title">
        <h2>投稿审核</h2>
        <span>{{ isAdminLoggedIn ? `管理员 ${adminUser?.username || ''}` : '请先登录管理员账号' }}</span>
      </div>

      <div v-if="!isAdminLoggedIn" class="login-row">
        <t-input
          v-model="adminForm.username"
          placeholder="管理员账号"
          :disabled="adminLoading"
        />
        <t-input
          v-model="adminForm.password"
          type="password"
          placeholder="管理员密码"
          :disabled="adminLoading"
          @keydown.enter.prevent="loginAdmin"
        />
        <t-button
          theme="primary"
          :loading="adminLoading"
          :disabled="!canLoginAdmin"
          @click="loginAdmin"
        >
          登录
        </t-button>
      </div>

      <div v-else class="review-layout">
        <aside class="submission-list">
          <div class="submission-toolbar">
            <select v-model="reviewStatus" class="status-select" @change="loadSubmissions">
              <option value="pending">待审核</option>
              <option value="approved">已通过</option>
              <option value="rejected">已驳回</option>
              <option value="">全部</option>
            </select>
            <t-button
              variant="outline"
              :loading="submissionsLoading"
              @click="loadSubmissions"
            >
              刷新
            </t-button>
            <t-button variant="text" @click="logoutAdmin">退出</t-button>
          </div>

          <div v-if="!submissions.length && !submissionsLoading" class="empty-state">
            暂无投稿
          </div>

          <button
            v-for="submission in submissions"
            :key="submission.id"
            class="submission-item"
            :class="{ active: selectedSubmission?.id === submission.id }"
            type="button"
            @click="fillReviewForm(submission)"
          >
            <strong>{{ submission.question }}</strong>
            <span>{{ submission.status }} · #{{ submission.id }}</span>
          </button>
        </aside>

        <section class="review-detail">
          <div v-if="selectedSubmission" class="review-form">
            <t-form label-align="top" @submit.prevent>
              <t-form-item label="问题">
                <t-textarea
                  v-model="reviewForm.question"
                  :autosize="{ minRows: 2, maxRows: 4 }"
                  :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                />
              </t-form-item>
              <t-form-item label="答案">
                <t-textarea
                  v-model="reviewForm.answer"
                  :autosize="{ minRows: 4, maxRows: 7 }"
                  :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                />
              </t-form-item>

              <div class="form-row">
                <t-form-item label="分类">
                  <t-input
                    v-model="reviewForm.category"
                    :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                  />
                </t-form-item>
                <t-form-item label="标签">
                  <t-input
                    v-model="reviewForm.tags"
                    placeholder="食堂，营业时间"
                    :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                  />
                </t-form-item>
              </div>

              <div class="form-row">
                <t-form-item label="来源">
                  <t-input
                    v-model="reviewForm.source"
                    :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                  />
                </t-form-item>
                <t-form-item label="备注">
                  <t-input
                    v-model="reviewForm.remark"
                    :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                  />
                </t-form-item>
              </div>

              <t-form-item label="审核备注">
                <t-input
                  v-model="reviewForm.reviewerNote"
                  placeholder="可填写通过或驳回原因"
                  :disabled="reviewLoading || selectedSubmission.status !== 'pending'"
                />
              </t-form-item>

              <div class="review-actions">
                <t-button
                  theme="success"
                  :loading="reviewLoading"
                  :disabled="!canApproveSubmission || selectedSubmission.status !== 'pending'"
                  @click="approveSubmission"
                >
                  通过并入库
                </t-button>
                <t-button
                  theme="danger"
                  variant="outline"
                  :loading="reviewLoading"
                  :disabled="!canRejectSubmission || selectedSubmission.status !== 'pending'"
                  @click="rejectSubmission"
                >
                  驳回
                </t-button>
              </div>
            </t-form>
          </div>

          <div v-else class="empty-state">请选择一条投稿</div>

          <t-alert
            v-if="adminMessage"
            class="feedback"
            theme="success"
            :message="adminMessage"
          />
          <t-alert
            v-if="adminError"
            class="feedback"
            theme="error"
            :message="adminError"
          />
          <t-alert
            v-if="reviewMessage"
            class="feedback"
            theme="success"
            :message="reviewMessage"
          />
          <t-alert
            v-if="reviewError"
            class="feedback"
            theme="error"
            :message="reviewError"
          />
        </section>
      </div>
    </section>

    <section class="console-grid">
      <section class="panel">
        <div class="panel-title">
          <h2>知识录入</h2>
          <span>保存时会生成向量并入库</span>
        </div>

        <t-form label-align="top" @submit.prevent>
          <t-form-item label="问题">
            <t-textarea
              v-model="knowledgeForm.question"
              placeholder="例如：三食堂晚上几点关门？"
              :autosize="{ minRows: 2, maxRows: 4 }"
              :disabled="submitLoading"
            />
          </t-form-item>

          <t-form-item label="答案">
            <t-textarea
              v-model="knowledgeForm.answer"
              placeholder="填写可以直接返回给用户的答案"
              :autosize="{ minRows: 5, maxRows: 8 }"
              :disabled="submitLoading"
            />
          </t-form-item>

          <div class="form-row">
            <t-form-item label="分类">
              <t-input
                v-model="knowledgeForm.category"
                placeholder="餐饮服务"
                :disabled="submitLoading"
              />
            </t-form-item>
            <t-form-item label="标签">
              <t-input
                v-model="knowledgeForm.tags"
                placeholder="食堂，营业时间，关门"
                :disabled="submitLoading"
              />
            </t-form-item>
          </div>

          <div class="form-row">
            <t-form-item label="来源">
              <t-input
                v-model="knowledgeForm.source"
                placeholder="后勤公告"
                :disabled="submitLoading"
              />
            </t-form-item>
            <t-form-item label="备注">
              <t-input
                v-model="knowledgeForm.remark"
                placeholder="可选"
                :disabled="submitLoading"
              />
            </t-form-item>
          </div>

          <t-button
            theme="primary"
            block
            :loading="submitLoading"
            :disabled="!canSubmitKnowledge"
            @click="submitKnowledge"
          >
            保存知识
          </t-button>
        </t-form>

        <t-alert
          v-if="submitMessage"
          class="feedback"
          theme="success"
          :message="submitMessage"
        />
        <t-alert
          v-if="submitError"
          class="feedback"
          theme="error"
          :message="submitError"
        />
      </section>

      <section class="panel">
        <div class="panel-title">
          <h2>问答测试</h2>
          <span>请求返回前会锁定提问区</span>
        </div>

        <t-textarea
          v-model="askForm.question"
          placeholder="例如：食堂几点关门？"
          :autosize="{ minRows: 4, maxRows: 6 }"
          :disabled="askLoading"
          @keydown.enter.prevent="askQuestion"
        />

        <div class="ask-actions">
          <t-button
            theme="primary"
            :loading="askLoading"
            :disabled="!canAsk"
            @click="askQuestion"
          >
            提问
          </t-button>
        </div>

        <t-alert
          v-if="askError"
          class="feedback"
          theme="error"
          :message="askError"
        />

        <t-alert
          v-if="askResult && !askResult.answered"
          class="feedback"
          theme="warning"
          :message="`未找到足够相关的答案。最高相似度 ${Number(askResult.candidates?.[0]?.score || 0).toFixed(4)}，命中阈值 ${Number(askResult.min_score || 0).toFixed(2)}。`"
        />

        <t-alert
          v-if="askResult?.ai_error"
          class="feedback"
          theme="warning"
          :message="`AI 回答生成失败，已返回知识库原始结果：${askResult.ai_error}`"
        />

        <div v-if="askResult?.ai_answer" class="ai-answer-box">
          <div class="answer-meta">
            <t-tag theme="primary" variant="light">AI 回答</t-tag>
            <span>基于候选知识生成</span>
          </div>
          <p>{{ askResult.ai_answer }}</p>
        </div>

        <div v-if="askResult?.answer" class="answer-box">
          <div class="answer-meta">
            <t-tag theme="success" variant="light">
              {{ askResult.answer.category || '未分类' }}
            </t-tag>
            <span>相似度 {{ Number(askResult.answer.score || 0).toFixed(4) }}</span>
          </div>
          <h3>{{ candidateTitle(askResult.answer) }}</h3>
          <p>{{ candidateBody(askResult.answer) }}</p>
          <a
            v-if="askResult.answer.source_url"
            class="source-link"
            :href="askResult.answer.source_url"
            target="_blank"
            rel="noreferrer"
          >
            查看来源
          </a>
          <div v-if="askResult.answer.tags?.length" class="tag-list">
            <t-tag
              v-for="tag in askResult.answer.tags"
              :key="tag"
              variant="light"
            >
              {{ tag }}
            </t-tag>
          </div>
        </div>

        <div v-if="askResult?.candidates?.length" class="candidate-section">
          <div class="candidate-title">
            <h3>候选结果</h3>
            <span>相似度 = 1 - 余弦距离，低于 {{ Number(askResult.min_score || 0).toFixed(2) }} 不自动回答</span>
          </div>
          <div class="candidate-list">
            <div
              v-for="(item, index) in askResult.candidates"
              :key="item.chunk_id || item.id"
              class="candidate-item"
            >
              <div class="candidate-rank">{{ index + 1 }}</div>
              <div class="candidate-body">
                <div class="candidate-head">
                  <strong>{{ candidateTitle(item) }}</strong>
                  <span>{{ Number(item.score || 0).toFixed(4) }}</span>
                </div>
                <p>{{ candidateBody(item) }}</p>
                <div class="candidate-foot">
                  <t-tag size="small" variant="light">
                    {{ item.category || '未分类' }}
                  </t-tag>
                  <span v-if="item.chunk_id">片段 #{{ item.chunk_id }}</span>
                  <span v-if="item.tags?.length">{{ item.tags.join(' / ') }}</span>
                  <a
                    v-if="item.source_url"
                    class="source-link"
                    :href="item.source_url"
                    target="_blank"
                    rel="noreferrer"
                  >
                    来源
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </section>
  </main>
</template>
