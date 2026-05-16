<script setup>
import { computed, reactive, ref } from 'vue'

const apiBaseURL = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080'

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

const submitLoading = ref(false)
const askLoading = ref(false)
const submitMessage = ref('')
const submitError = ref('')
const askError = ref('')
const askResult = ref(null)

const canSubmitKnowledge = computed(() => (
  knowledgeForm.question.trim() &&
  knowledgeForm.answer.trim() &&
  !submitLoading.value
))

const canAsk = computed(() => askForm.question.trim() && !askLoading.value)

function parseTags(value) {
  return value
    .split(/[,，\n]/)
    .map((tag) => tag.trim())
    .filter(Boolean)
}

async function postJSON(path, payload) {
  const response = await fetch(`${apiBaseURL}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })
  const result = await response.json().catch(() => null)
  if (!response.ok || result?.code !== 0) {
    throw new Error(result?.message || `HTTP ${response.status}`)
  }
  return result.data
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
</script>

<template>
  <main class="console-page">
    <header class="console-header">
      <div>
        <h1>校园生活百事通 Console</h1>
        <p>录入知识后可直接在右侧问答中检索验证。</p>
      </div>
      <t-tag theme="primary" variant="light">API {{ apiBaseURL }}</t-tag>
    </header>

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

        <div v-if="askResult?.answer" class="answer-box">
          <div class="answer-meta">
            <t-tag theme="success" variant="light">
              {{ askResult.answer.category || '未分类' }}
            </t-tag>
            <span>相似度 {{ Number(askResult.answer.score || 0).toFixed(4) }}</span>
          </div>
          <h3>{{ askResult.answer.matched_question }}</h3>
          <p>{{ askResult.answer.answer }}</p>
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
              :key="item.id"
              class="candidate-item"
            >
              <div class="candidate-rank">{{ index + 1 }}</div>
              <div class="candidate-body">
                <div class="candidate-head">
                  <strong>{{ item.matched_question }}</strong>
                  <span>{{ Number(item.score || 0).toFixed(4) }}</span>
                </div>
                <p>{{ item.answer }}</p>
                <div class="candidate-foot">
                  <t-tag size="small" variant="light">
                    {{ item.category || '未分类' }}
                  </t-tag>
                  <span v-if="item.tags?.length">{{ item.tags.join(' / ') }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </section>
  </main>
</template>
