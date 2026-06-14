<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { ChevronDownIcon } from 'tdesign-icons-vue-next'
import {
  CreateSubmission,
  GetCurrentUser,
  GetHotQuestionsStatus,
  ListMySubmissions,
  Logout,
} from '../../wailsjs/go/main/App'
import Chat from './Chat.vue'
import { errorMessage, isAuthExpiredError } from '../utils/errors'
import { formatDateTime, parseTags, submissionStatusText } from '../utils/student'

const props = defineProps({
  initialUser: {
    type: Object,
    default: () => null,
  },
})

const emit = defineEmits(['logout'])

const navItems = [
  { id: 'ask', label: '问答' },
  { id: 'submit', label: '贡献问答' },
  { id: 'history', label: '我的投稿' },
  { id: 'hot', label: '热点问题' },
  { id: 'profile', label: '个人信息' },
]

const activeView = ref('ask')
const showDropdown = ref(false)
const shellError = ref('')
const shellNotice = ref('')
const user = ref({
  username: props.initialUser?.username || '',
  nickname: props.initialUser?.nickname || '',
})

const profileLoading = ref(false)
const submitLoading = ref(false)
const historyLoading = ref(false)
const hotLoading = ref(false)

const submitMessage = ref('')
const submitError = ref('')
const historyError = ref('')
const hotState = ref(null)
const submissions = ref([])
const selectedSubmissionID = ref(null)

const submissionForm = reactive({
  question: '',
  answer: '',
  category: '',
  tags: '',
  source: '',
  remark: '',
})

const selectedSubmission = computed(() => (
  submissions.value.find((item) => item.id === selectedSubmissionID.value) || null
))

const displayName = computed(() => (
  user.value.nickname || user.value.username || '用户'
))

const submitDisabled = computed(() => (
  submitLoading.value ||
  !submissionForm.question.trim() ||
  !submissionForm.answer.trim()
))

function resetShellFeedback() {
  shellError.value = ''
  shellNotice.value = ''
}

function switchView(viewID) {
  activeView.value = viewID
  showDropdown.value = false
}

function closeDropdown(event) {
  if (!event.target.closest('.shell-user-menu')) {
    showDropdown.value = false
  }
}

function handleSessionExpired() {
  emit('logout')
}

function handleProtectedError(error, fallback) {
  const message = errorMessage(error, fallback)
  if (isAuthExpiredError(error)) {
    handleSessionExpired()
    return true
  }
  shellError.value = message
  return false
}

async function refreshProfile(options = {}) {
  const { quiet = false } = options
  profileLoading.value = true
  if (!quiet) {
    resetShellFeedback()
  }

  try {
    const profile = await GetCurrentUser()
    user.value = profile || user.value
  } catch (error) {
    if (!quiet) {
      handleProtectedError(error, '获取个人信息失败')
    } else if (isAuthExpiredError(error)) {
      handleSessionExpired()
    }
  } finally {
    profileLoading.value = false
  }
}

async function submitContribution() {
  if (submitDisabled.value) return

  submitLoading.value = true
  submitError.value = ''
  submitMessage.value = ''
  shellError.value = ''
  shellNotice.value = ''

  try {
    await CreateSubmission({
      question: submissionForm.question.trim(),
      answer: submissionForm.answer.trim(),
      category: submissionForm.category.trim(),
      tags: parseTags(submissionForm.tags),
      source: submissionForm.source.trim(),
      remark: submissionForm.remark.trim(),
    })
    submitMessage.value = '投稿已提交，等待管理员审核。'
    shellNotice.value = '投稿提交成功'
    submissionForm.question = ''
    submissionForm.answer = ''
    submissionForm.category = ''
    submissionForm.tags = ''
    submissionForm.source = ''
    submissionForm.remark = ''
    await loadSubmissionHistory({ quiet: true })
  } catch (error) {
    if (!handleProtectedError(error, '投稿提交失败')) {
      submitError.value = errorMessage(error)
    }
  } finally {
    submitLoading.value = false
  }
}

async function loadSubmissionHistory(options = {}) {
  const { quiet = false } = options
  historyLoading.value = true
  if (!quiet) {
    historyError.value = ''
    shellError.value = ''
  }

  try {
    const result = await ListMySubmissions()
    submissions.value = Array.isArray(result) ? result : []
    selectedSubmissionID.value = submissions.value[0]?.id || null
  } catch (error) {
    if (!handleProtectedError(error, '获取投稿历史失败')) {
      historyError.value = errorMessage(error)
    }
  } finally {
    historyLoading.value = false
  }
}

async function loadHotQuestionsStatus() {
  if (hotLoading.value) return

  hotLoading.value = true
  shellError.value = ''

  try {
    hotState.value = await GetHotQuestionsStatus(10)
  } catch (error) {
    if (!handleProtectedError(error, '获取热点问题状态失败')) {
      hotState.value = {
        available: false,
        message: errorMessage(error),
      }
    }
  } finally {
    hotLoading.value = false
  }
}

async function logoutStudent() {
  shellError.value = ''
  shellNotice.value = ''

  try {
    await Logout()
    emit('logout')
  } catch (error) {
    if (isAuthExpiredError(error)) {
      emit('logout')
      return
    }
    shellError.value = errorMessage(error, '退出登录失败')
  }
}

watch(activeView, async (viewID) => {
  if (viewID === 'history' && submissions.value.length === 0) {
    await loadSubmissionHistory()
  }
  if (viewID === 'profile' && !profileLoading.value) {
    await refreshProfile({ quiet: true })
  }
  if (viewID === 'hot' && hotState.value == null) {
    await loadHotQuestionsStatus()
  }
})

onMounted(async () => {
  document.addEventListener('click', closeDropdown)
  await refreshProfile()
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown)
})
</script>

<template>
  <div class="shell-root">
    <header class="shell-header">
      <div class="shell-brand">
        <div class="shell-brand-mark">问</div>
        <div>
          <div class="shell-brand-title">校园生活百事通</div>
          <div class="shell-brand-subtitle">学生桌面客户端</div>
        </div>
      </div>

      <nav class="shell-nav">
        <button
          v-for="item in navItems"
          :key="item.id"
          class="shell-nav-btn"
          :class="{ active: activeView === item.id }"
          @click="switchView(item.id)"
        >
          {{ item.label }}
        </button>
      </nav>

      <div class="shell-user-menu" @click.stop="showDropdown = !showDropdown">
        <div class="shell-user-avatar">{{ displayName.charAt(0) }}</div>
        <div class="shell-user-meta">
          <strong>{{ displayName }}</strong>
          <span>{{ user.username || '未命名账号' }}</span>
        </div>
        <ChevronDownIcon class="shell-user-arrow" :class="{ open: showDropdown }" />

        <div v-if="showDropdown" class="shell-dropdown">
          <button class="shell-dropdown-item" @click.stop="switchView('profile')">查看个人信息</button>
          <button class="shell-dropdown-item" @click.stop="switchView('history')">查看我的投稿</button>
          <button class="shell-dropdown-item danger" @click.stop="logoutStudent">退出登录</button>
        </div>
      </div>
    </header>

    <main class="shell-main">
      <aside class="shell-side-card">
        <h2>欢迎回来</h2>
        <p>{{ displayName }}，问答、投稿和审核进度都在这里。</p>
        <dl class="shell-side-list">
          <div>
            <dt>登录身份</dt>
            <dd>学生</dd>
          </div>
          <div>
            <dt>账号</dt>
            <dd>{{ user.username || '-' }}</dd>
          </div>
          <div>
            <dt>当前入口</dt>
            <dd>{{ navItems.find((item) => item.id === activeView)?.label }}</dd>
          </div>
        </dl>
      </aside>

      <section class="shell-content">
        <p v-if="shellError" class="shell-banner error">{{ shellError }}</p>
        <p v-else-if="shellNotice" class="shell-banner success">{{ shellNotice }}</p>

        <div v-show="activeView === 'ask'" class="panel-wrap">
          <Chat
            :user-name="displayName"
            @session-expired="handleSessionExpired"
          />
        </div>

        <div v-show="activeView === 'submit'" class="panel-wrap">
          <section class="panel-card">
            <div class="panel-head">
              <div>
                <h2>贡献问答</h2>
                <p>提交你确认过的校园信息，管理员审核通过后会进入知识库。</p>
              </div>
            </div>

            <div class="form-grid">
              <label class="field wide">
                <span>问题</span>
                <input v-model="submissionForm.question" type="text" placeholder="例如：图书馆周末几点闭馆？">
              </label>

              <label class="field wide">
                <span>答案</span>
                <textarea v-model="submissionForm.answer" rows="5" placeholder="填写清晰、可验证的回答内容" />
              </label>

              <label class="field">
                <span>分类</span>
                <input v-model="submissionForm.category" type="text" placeholder="例如：图书馆">
              </label>

              <label class="field">
                <span>标签</span>
                <input v-model="submissionForm.tags" type="text" placeholder="用逗号分隔多个标签">
              </label>

              <label class="field wide">
                <span>来源</span>
                <input v-model="submissionForm.source" type="text" placeholder="官网链接、通知标题或线下窗口来源">
              </label>

              <label class="field wide">
                <span>备注</span>
                <textarea v-model="submissionForm.remark" rows="3" placeholder="可选，补充范围、日期或注意事项" />
              </label>
            </div>

            <p v-if="submitError" class="inline-feedback error">{{ submitError }}</p>
            <p v-else-if="submitMessage" class="inline-feedback success">{{ submitMessage }}</p>

            <div class="panel-actions">
              <button class="primary-btn" :disabled="submitDisabled" @click="submitContribution">
                {{ submitLoading ? '提交中...' : '提交投稿' }}
              </button>
            </div>
          </section>
        </div>

        <div v-show="activeView === 'history'" class="panel-wrap">
          <section class="panel-card">
            <div class="panel-head">
              <div>
                <h2>我的投稿</h2>
                <p>查看你提交过的问题、审核状态和管理员反馈。</p>
              </div>
              <button class="secondary-btn" :disabled="historyLoading" @click="loadSubmissionHistory">
                {{ historyLoading ? '刷新中...' : '刷新列表' }}
              </button>
            </div>

            <p v-if="historyError" class="inline-feedback error">{{ historyError }}</p>
            <div v-else-if="!historyLoading && submissions.length === 0" class="empty-state">
              暂无投稿记录。
            </div>
            <div v-else class="history-layout">
              <div class="history-list">
                <button
                  v-for="item in submissions"
                  :key="item.id"
                  class="history-item"
                  :class="{ active: selectedSubmissionID === item.id }"
                  @click="selectedSubmissionID = item.id"
                >
                  <div class="history-item-head">
                    <strong>{{ item.question }}</strong>
                    <span class="status-pill" :class="item.status">{{ submissionStatusText(item.status) }}</span>
                  </div>
                  <p>{{ item.answer }}</p>
                  <small>{{ formatDateTime(item.created_at) }}</small>
                </button>
              </div>

              <article v-if="selectedSubmission" class="history-detail">
                <h3>{{ selectedSubmission.question }}</h3>
                <p>{{ selectedSubmission.answer }}</p>
                <dl>
                  <div>
                    <dt>分类</dt>
                    <dd>{{ selectedSubmission.category || '-' }}</dd>
                  </div>
                  <div>
                    <dt>标签</dt>
                    <dd>{{ selectedSubmission.tags?.join('、') || '-' }}</dd>
                  </div>
                  <div>
                    <dt>来源</dt>
                    <dd>{{ selectedSubmission.source || '-' }}</dd>
                  </div>
                  <div>
                    <dt>备注</dt>
                    <dd>{{ selectedSubmission.remark || '-' }}</dd>
                  </div>
                  <div>
                    <dt>审核状态</dt>
                    <dd>{{ submissionStatusText(selectedSubmission.status) }}</dd>
                  </div>
                  <div>
                    <dt>审核意见</dt>
                    <dd>{{ selectedSubmission.reviewer_note || '暂无' }}</dd>
                  </div>
                  <div>
                    <dt>提交时间</dt>
                    <dd>{{ formatDateTime(selectedSubmission.created_at) }}</dd>
                  </div>
                  <div>
                    <dt>审核时间</dt>
                    <dd>{{ formatDateTime(selectedSubmission.reviewed_at) }}</dd>
                  </div>
                </dl>
              </article>
            </div>
          </section>
        </div>

        <div v-show="activeView === 'hot'" class="panel-wrap">
          <section class="panel-card">
            <div class="panel-head">
              <div>
                <h2>热点问题</h2>
                <p>该入口已经对接后端状态，当前按服务实际能力展示。</p>
              </div>
              <button class="secondary-btn" :disabled="hotLoading" @click="loadHotQuestionsStatus">
                {{ hotLoading ? '检测中...' : '重新检测' }}
              </button>
            </div>

            <div v-if="hotLoading" class="empty-state">正在获取热点问题状态...</div>
            <div v-else-if="hotState?.available" class="empty-state">
              热点问题接口已可用，当前客户端已完成状态对齐，可继续扩展列表展示。
            </div>
            <div v-else class="empty-state">
              {{ hotState?.message || '热点问题功能开发中。' }}
            </div>
          </section>
        </div>

        <div v-show="activeView === 'profile'" class="panel-wrap">
          <section class="panel-card">
            <div class="panel-head">
              <div>
                <h2>个人信息</h2>
                <p>此处数据直接来自 `/api/v1/users/me`。</p>
              </div>
              <button class="secondary-btn" :disabled="profileLoading" @click="refreshProfile">
                {{ profileLoading ? '加载中...' : '刷新信息' }}
              </button>
            </div>

            <dl class="profile-grid">
              <div>
                <dt>昵称</dt>
                <dd>{{ user.nickname || '-' }}</dd>
              </div>
              <div>
                <dt>账号</dt>
                <dd>{{ user.username || '-' }}</dd>
              </div>
              <div>
                <dt>身份</dt>
                <dd>学生</dd>
              </div>
            </dl>
          </section>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.shell-root {
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.shell-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 20px;
  align-items: center;
  padding: 18px 28px;
  background: rgba(255, 255, 255, 0.86);
  backdrop-filter: blur(18px);
  border-bottom: 1px solid rgba(148, 163, 184, 0.20);
}

.shell-brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.shell-brand-mark {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  font-size: 18px;
  font-weight: 800;
  color: #fff;
  background: linear-gradient(135deg, #f97316 0%, #0ea5e9 100%);
}

.shell-brand-title {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
}

.shell-brand-subtitle {
  font-size: 12px;
  color: #64748b;
}

.shell-nav {
  display: flex;
  gap: 10px;
  justify-content: center;
  flex-wrap: wrap;
}

.shell-nav-btn,
.secondary-btn,
.primary-btn,
.shell-dropdown-item,
.history-item {
  font: inherit;
}

.shell-nav-btn {
  border: none;
  border-radius: 999px;
  padding: 10px 16px;
  color: #475569;
  background: transparent;
  cursor: pointer;
  transition: 160ms ease;
}

.shell-nav-btn:hover,
.shell-nav-btn.active {
  color: #0f172a;
  background: rgba(14, 165, 233, 0.10);
}

.shell-user-menu {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 18px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.76);
  border: 1px solid rgba(226, 232, 240, 0.9);
}

.shell-user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  background: linear-gradient(135deg, #fb923c 0%, #38bdf8 100%);
}

.shell-user-meta {
  display: flex;
  flex-direction: column;
}

.shell-user-meta strong {
  font-size: 14px;
  color: #0f172a;
}

.shell-user-meta span {
  font-size: 12px;
  color: #64748b;
}

.shell-user-arrow {
  color: #64748b;
  transition: transform 160ms ease;
}

.shell-user-arrow.open {
  transform: rotate(180deg);
}

.shell-dropdown {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 180px;
  padding: 8px;
  border-radius: 18px;
  background: #fff;
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 0 20px 40px rgba(15, 23, 42, 0.08);
  z-index: 20;
}

.shell-dropdown-item {
  width: 100%;
  border: none;
  background: transparent;
  border-radius: 12px;
  padding: 10px 12px;
  text-align: left;
  cursor: pointer;
  color: #334155;
}

.shell-dropdown-item:hover {
  background: rgba(14, 165, 233, 0.08);
}

.shell-dropdown-item.danger {
  color: #b91c1c;
}

.shell-main {
  flex: 1;
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 24px;
  padding: 24px 28px 28px;
}

.shell-side-card,
.panel-card {
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(226, 232, 240, 0.95);
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.07);
}

.shell-side-card {
  padding: 24px;
  height: fit-content;
}

.shell-side-card h2 {
  margin: 0 0 8px;
  font-size: 24px;
  color: #0f172a;
}

.shell-side-card p {
  margin: 0 0 20px;
  color: #475569;
  line-height: 1.6;
}

.shell-side-list {
  display: grid;
  gap: 16px;
  margin: 0;
}

.shell-side-list div,
.profile-grid div,
.history-detail dl div {
  display: grid;
  gap: 6px;
}

.shell-side-list dt,
.profile-grid dt,
.history-detail dt {
  font-size: 12px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #94a3b8;
}

.shell-side-list dd,
.profile-grid dd,
.history-detail dd {
  margin: 0;
  color: #0f172a;
  font-weight: 700;
}

.shell-content {
  min-width: 0;
}

.shell-banner {
  margin: 0 0 16px;
  padding: 14px 18px;
  border-radius: 18px;
}

.shell-banner.error,
.inline-feedback.error {
  color: #991b1b;
  background: #fef2f2;
}

.shell-banner.success,
.inline-feedback.success {
  color: #166534;
  background: #f0fdf4;
}

.panel-wrap {
  min-width: 0;
}

.panel-card {
  padding: 24px;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.panel-head h2 {
  margin: 0 0 6px;
  color: #0f172a;
}

.panel-head p {
  margin: 0;
  color: #64748b;
}

.secondary-btn,
.primary-btn {
  border: none;
  border-radius: 14px;
  cursor: pointer;
  transition: 160ms ease;
}

.secondary-btn {
  padding: 10px 14px;
  background: #e2e8f0;
  color: #334155;
}

.secondary-btn:hover,
.primary-btn:hover {
  transform: translateY(-1px);
}

.primary-btn {
  padding: 12px 18px;
  background: linear-gradient(135deg, #f97316 0%, #0ea5e9 100%);
  color: #fff;
  font-weight: 700;
}

.secondary-btn:disabled,
.primary-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
  transform: none;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.field {
  display: grid;
  gap: 8px;
}

.field.wide {
  grid-column: 1 / -1;
}

.field span {
  color: #334155;
  font-weight: 700;
}

.field input,
.field textarea {
  width: 100%;
  border: 1px solid #dbe4ee;
  border-radius: 16px;
  padding: 14px 16px;
  background: #fff;
  color: #0f172a;
  resize: vertical;
}

.field input:focus,
.field textarea:focus {
  outline: none;
  border-color: #38bdf8;
  box-shadow: 0 0 0 4px rgba(56, 189, 248, 0.14);
}

.inline-feedback {
  margin: 16px 0 0;
  padding: 12px 14px;
  border-radius: 14px;
}

.panel-actions {
  margin-top: 18px;
  display: flex;
  justify-content: flex-end;
}

.history-layout {
  display: grid;
  grid-template-columns: minmax(280px, 340px) minmax(0, 1fr);
  gap: 18px;
}

.history-list {
  display: grid;
  gap: 12px;
}

.history-item {
  border: 1px solid #e2e8f0;
  border-radius: 18px;
  padding: 16px;
  text-align: left;
  background: #fff;
  cursor: pointer;
  transition: 160ms ease;
}

.history-item:hover,
.history-item.active {
  border-color: #38bdf8;
  box-shadow: 0 12px 30px rgba(14, 165, 233, 0.08);
}

.history-item-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.history-item p {
  margin: 10px 0;
  color: #475569;
  line-height: 1.5;
}

.history-item small {
  color: #94a3b8;
}

.history-detail {
  border-radius: 22px;
  padding: 20px;
  background: #fff;
  border: 1px solid #e2e8f0;
}

.history-detail h3 {
  margin: 0 0 10px;
  color: #0f172a;
}

.history-detail > p {
  margin: 0 0 20px;
  line-height: 1.7;
  color: #475569;
}

.history-detail dl,
.profile-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin: 0;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}

.status-pill.pending {
  background: #fff7ed;
  color: #c2410c;
}

.status-pill.approved {
  background: #f0fdf4;
  color: #166534;
}

.status-pill.rejected {
  background: #fef2f2;
  color: #b91c1c;
}

.empty-state {
  border-radius: 18px;
  padding: 24px;
  background: #f8fafc;
  color: #475569;
}

@media (max-width: 1180px) {
  .shell-header,
  .shell-main,
  .history-layout,
  .profile-grid,
  .history-detail dl,
  .form-grid {
    grid-template-columns: 1fr;
  }

  .shell-header {
    justify-items: stretch;
  }

  .shell-user-menu {
    justify-self: end;
  }
}

@media (max-width: 768px) {
  .shell-header,
  .shell-main {
    padding: 18px;
  }

  .panel-card,
  .shell-side-card {
    border-radius: 22px;
  }

  .panel-head {
    flex-direction: column;
  }
}
</style>
