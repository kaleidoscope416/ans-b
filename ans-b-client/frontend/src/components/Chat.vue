<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { SendIcon } from 'tdesign-icons-vue-next'
import { AskQuestion } from '../../wailsjs/go/main/App'
import { errorMessage, isAuthExpiredError } from '../utils/errors'
import { replaceMessage } from '../utils/messages'
import assistantAvatar from '../assets/images/assistant-xiaobai.jpg'

const props = defineProps({
  userName: {
    type: String,
    default: '用户',
  },
})

const emit = defineEmits(['session-expired'])

const messages = ref([
  {
    id: 1,
    role: 'ai',
    content: '你好，我是小白。食堂、图书馆、校园卡、宿舍报修这些校园生活问题，都可以直接问我。',
    time: '10:30',
  },
])

const inputText = ref('')
const msgListRef = ref(null)
const isSending = ref(false)

const isInputEmpty = computed(() => !inputText.value.trim() || isSending.value)

function getCurrentTime() {
  const now = new Date()
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`
}

function formatScore(score) {
  const value = Number(score || 0)
  return value ? value.toFixed(4) : ''
}

function getCandidateTitle(item) {
  return item?.title || item?.matched_question || `知识 #${item?.item_id || item?.id || ''}`
}

function getCandidateBody(item) {
  return item?.chunk_text || item?.answer || ''
}

function getAnswerText(result) {
  return result?.ai_answer || getCandidateBody(result?.answer) || ''
}

function buildAnswerContent(result) {
  const answer = getAnswerText(result)
  if (answer) return answer
  if (result && result.answered === false) {
    return '未找到足够相关的答案。你可以换一种问法，或者补充更具体的地点、时间、业务名称。'
  }
  return '接口已返回，但没有可展示的回答内容。'
}

function answerMeta(result) {
  const answer = result?.answer
  if (!answer) return []
  return [
    answer.category || '未分类',
    formatScore(answer.score) ? `相似度 ${formatScore(answer.score)}` : '',
  ].filter(Boolean)
}

function candidateKey(item, index) {
  return item?.chunk_id || item?.id || `${item?.item_id || 'candidate'}-${index}`
}

function topCandidates(result) {
  return (result?.candidates || []).slice(0, 3)
}

async function scrollToBottom() {
  await nextTick()
  if (msgListRef.value) {
    msgListRef.value.scrollTop = msgListRef.value.scrollHeight
  }
}

async function handleSend(prefilledQuestion = '') {
  const text = (prefilledQuestion || inputText.value).trim()
  if (!text || isSending.value) return

  messages.value.push({
    id: Date.now(),
    role: 'user',
    content: text,
    time: getCurrentTime(),
  })
  inputText.value = ''
  await scrollToBottom()

  const pendingMessage = {
    id: Date.now() + 1,
    role: 'ai',
    content: '正在检索校园知识库...',
    time: getCurrentTime(),
    loading: true,
  }
  messages.value.push(pendingMessage)
  isSending.value = true
  await scrollToBottom()

  try {
    const result = await AskQuestion(text, 5)
    messages.value = replaceMessage(messages.value, pendingMessage.id, {
      content: buildAnswerContent(result),
      result,
      aiError: result?.ai_error || '',
      loading: false,
      time: getCurrentTime(),
    })
  } catch (error) {
    if (isAuthExpiredError(error)) {
      emit('session-expired')
    }
    messages.value = replaceMessage(messages.value, pendingMessage.id, {
      content: `问答接口请求失败：${errorMessage(error)}`,
      error: true,
      loading: false,
      time: getCurrentTime(),
    })
  } finally {
    isSending.value = false
    await scrollToBottom()
  }
}

function handleKeydown(event) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    handleSend()
  }
}

function askCandidate(question) {
  handleSend(question)
}

onMounted(() => {
  scrollToBottom()
})
</script>

<template>
  <section class="chat-card">
    <div class="chat-hero">
      <div>
        <h2>校园智能问答</h2>
        <p>直接提问，系统会先检索知识库，再在可用时返回 AI 增强答案。</p>
      </div>
      <div class="chat-user-chip">
        <span>{{ props.userName.charAt(0) }}</span>
        <strong>{{ props.userName }}</strong>
      </div>
    </div>

    <div ref="msgListRef" class="msg-list">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="msg-row"
        :class="msg.role"
      >
        <div class="msg-avatar">
          <img v-if="msg.role === 'ai'" :src="assistantAvatar" alt="小白">
          <span v-else>{{ props.userName.charAt(0) }}</span>
        </div>

        <div class="msg-body">
          <div class="msg-bubble" :class="{ error: msg.error, loading: msg.loading }">
            <p class="msg-content">{{ msg.content }}</p>

            <div v-if="msg.result" class="answer-details">
              <div v-if="answerMeta(msg.result).length" class="answer-meta-row">
                <span v-for="meta in answerMeta(msg.result)" :key="meta" class="meta-pill">
                  {{ meta }}
                </span>
              </div>

              <p v-if="msg.aiError" class="ai-error">
                AI 回答生成失败，已展示知识库原始结果：{{ msg.aiError }}
              </p>

              <div v-if="msg.result.answer" class="matched-card">
                <strong>{{ getCandidateTitle(msg.result.answer) }}</strong>
                <p>{{ getCandidateBody(msg.result.answer) }}</p>
              </div>

              <div v-if="topCandidates(msg.result).length" class="candidate-panel">
                <div class="candidate-title">相关候选</div>
                <button
                  v-for="(item, index) in topCandidates(msg.result)"
                  :key="candidateKey(item, index)"
                  class="candidate-item"
                  @click="askCandidate(getCandidateTitle(item))"
                >
                  <span>{{ index + 1 }}</span>
                  <strong>{{ getCandidateTitle(item) }}</strong>
                  <em>{{ formatScore(item.score) }}</em>
                </button>
              </div>
            </div>
          </div>
          <small class="msg-time">{{ msg.time }}</small>
        </div>
      </div>
    </div>

    <div class="chat-input-area">
      <input
        v-model="inputText"
        type="text"
        class="chat-input"
        :disabled="isSending"
        placeholder="输入问题，按 Enter 发送"
        @keydown="handleKeydown"
      >
      <button
        class="send-btn"
        :disabled="isInputEmpty"
        @click="handleSend()"
      >
        <SendIcon class="send-icon" />
        <span>{{ isSending ? '发送中' : '发送' }}</span>
      </button>
    </div>
  </section>
</template>

<style scoped>
.chat-card {
  border-radius: 28px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(226, 232, 240, 0.95);
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.07);
}

.chat-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.chat-hero h2 {
  margin: 0 0 6px;
  color: #0f172a;
}

.chat-hero p {
  margin: 0;
  color: #64748b;
}

.chat-user-chip {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 999px;
  background: #fff7ed;
  color: #9a3412;
}

.chat-user-chip span {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #f97316 0%, #0ea5e9 100%);
  color: #fff;
}

.msg-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 420px;
  max-height: 62vh;
  overflow-y: auto;
  padding-right: 8px;
}

.msg-row {
  display: flex;
  gap: 14px;
}

.msg-row.user {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, #f97316 0%, #0ea5e9 100%);
  color: #fff;
  font-weight: 700;
}

.msg-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.msg-body {
  max-width: min(78%, 720px);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.msg-row.user .msg-body {
  align-items: flex-end;
}

.msg-bubble {
  border-radius: 22px;
  padding: 16px 18px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.04);
}

.msg-row.user .msg-bubble {
  background: linear-gradient(135deg, #fff7ed 0%, #eff6ff 100%);
}

.msg-bubble.error {
  border-color: #fecaca;
  background: #fef2f2;
}

.msg-bubble.loading {
  opacity: 0.8;
}

.msg-content {
  margin: 0;
  line-height: 1.7;
  color: #1e293b;
  white-space: pre-wrap;
}

.msg-time {
  color: #94a3b8;
}

.answer-details {
  margin-top: 14px;
  display: grid;
  gap: 12px;
}

.answer-meta-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  color: #0f766e;
  background: #ecfeff;
}

.ai-error {
  margin: 0;
  color: #b45309;
  background: #fffbeb;
  border-radius: 14px;
  padding: 10px 12px;
}

.matched-card {
  border-radius: 18px;
  padding: 14px 16px;
  background: #f8fafc;
}

.matched-card strong,
.candidate-item strong {
  color: #0f172a;
}

.matched-card p {
  margin: 8px 0 0;
  color: #475569;
  line-height: 1.6;
}

.candidate-panel {
  display: grid;
  gap: 10px;
}

.candidate-title {
  color: #334155;
  font-weight: 700;
}

.candidate-item {
  width: 100%;
  display: grid;
  grid-template-columns: 28px 1fr auto;
  gap: 10px;
  align-items: center;
  border: 1px solid #dbe4ee;
  border-radius: 16px;
  padding: 12px 14px;
  background: #fff;
  cursor: pointer;
}

.candidate-item span {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #0f766e;
  background: #ecfeff;
  font-weight: 700;
}

.candidate-item em {
  font-style: normal;
  color: #64748b;
}

.chat-input-area {
  margin-top: 20px;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
  align-items: center;
}

.chat-input {
  width: 100%;
  border: 1px solid #dbe4ee;
  border-radius: 18px;
  padding: 16px 18px;
  background: #fff;
}

.chat-input:focus {
  outline: none;
  border-color: #38bdf8;
  box-shadow: 0 0 0 4px rgba(56, 189, 248, 0.14);
}

.send-btn {
  border: none;
  border-radius: 18px;
  padding: 14px 18px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(135deg, #f97316 0%, #0ea5e9 100%);
  font-weight: 700;
}

.send-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .chat-hero,
  .chat-input-area {
    grid-template-columns: 1fr;
    display: grid;
  }

  .msg-body {
    max-width: 100%;
  }
}
</style>
