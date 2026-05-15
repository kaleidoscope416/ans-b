<script setup>
import { ref, reactive, nextTick, computed, onMounted, onUnmounted } from 'vue'
import { AttachIcon, SendIcon, ChevronDownIcon } from 'tdesign-icons-vue-next'

const props = defineProps({
  userName: {
    type: String,
    default: '用户',
  },
})

const emit = defineEmits(['logout'])

const messages = ref([
  {
    id: 1,
    role: 'ai',
    content: '你好！我是你的智能助手，有什么可以帮你的？',
    time: '10:30',
  },
])

const inputText = ref('')
const msgListRef = ref(null)
const showDropdown = ref(false)

const isInputEmpty = computed(() => !inputText.value.trim())

async function scrollToBottom() {
  await nextTick()
  if (msgListRef.value) {
    msgListRef.value.scrollTop = msgListRef.value.scrollHeight
  }
}

function getCurrentTime() {
  const now = new Date()
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`
}

function handleSend() {
  const text = inputText.value.trim()
  if (!text) return

  // 用户消息
  messages.value.push({
    id: Date.now(),
    role: 'user',
    content: text,
    time: getCurrentTime(),
  })
  inputText.value = ''
  scrollToBottom()

  // Mock AI 回复
  setTimeout(() => {
    messages.value.push({
      id: Date.now() + 1,
      role: 'ai',
      content: `收到你的问题："${text}"\n\n这是一个 Mock 回复。在实际项目中，这里会调用后端 AI 接口返回真实回答。`,
      time: getCurrentTime(),
    })
    scrollToBottom()
  }, 600)
}

function handleKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

function closeDropdown(e) {
  if (!e.target.closest('.user-menu')) {
    showDropdown.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', closeDropdown)
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown)
})

function handleLogout() {
  showDropdown.value = false
  emit('logout')
}
</script>

<template>
  <div class="chat-container">
    <!-- 顶部导航栏 -->
    <header class="chat-header">
      <nav class="header-left">
        <button class="nav-btn">贡献问答</button>
        <button class="nav-btn">热点问题</button>
      </nav>
      <div class="header-right">
        <div class="user-menu" @click.stop="showDropdown = !showDropdown">
          <div class="user-avatar">
            <img
              v-if="false"
              src=""
              alt="avatar"
            />
            <div v-else class="avatar-fallback">{{ userName.charAt(0) }}</div>
          </div>
          <span class="user-name">{{ userName }}</span>
          <ChevronDownIcon class="dropdown-arrow" :class="{ open: showDropdown }" />
          <Transition name="dropdown">
            <div v-if="showDropdown" class="dropdown-menu">
              <button class="dropdown-item">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  <circle cx="12" cy="7" r="4"/>
                </svg>
                个人信息
              </button>
              <div class="dropdown-divider" />
              <button class="dropdown-item danger" @click.stop="handleLogout">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                  <polyline points="16 17 21 12 16 7"/>
                  <line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                退出登录
              </button>
            </div>
          </Transition>
        </div>
      </div>
    </header>

    <!-- 消息区域 -->
    <div ref="msgListRef" class="msg-list">
      <TransitionGroup name="msg-bubble">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="msg-row"
          :class="msg.role"
        >
          <div class="msg-avatar">
            <div v-if="msg.role === 'ai'" class="ai-avatar">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="3" y="11" width="18" height="10" rx="2"/>
                <circle cx="12" cy="5" r="2"/>
                <path d="M12 7v4"/>
                <line x1="8" y1="16" x2="8" y2="16"/>
                <line x1="16" y1="16" x2="16" y2="16"/>
              </svg>
            </div>
            <div v-else class="user-avatar-small">
              {{ userName.charAt(0) }}
            </div>
          </div>
          <div class="msg-body">
            <div class="msg-bubble">
              {{ msg.content }}
            </div>
            <span class="msg-time">{{ msg.time }}</span>
          </div>
        </div>
      </TransitionGroup>
    </div>

    <!-- 底部输入区 -->
    <div class="chat-input-area">
      <button class="attach-btn" title="添加附件">
        <AttachIcon class="attach-icon" />
      </button>
      <div class="input-wrapper">
        <input
          v-model="inputText"
          type="text"
          placeholder="输入问题，按 Enter 发送"
          class="chat-input"
          @keydown="handleKeydown"
        />
      </div>
      <button
        class="send-btn"
        :class="{ disabled: isInputEmpty }"
        :disabled="isInputEmpty"
        @click="handleSend"
      >
        <SendIcon class="send-icon" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100vh;
  background: #F8FAFC;
}

/* ── 顶部导航栏 ── */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 24px;
  background: #FFFFFF;
  box-shadow: 0 1px 0 rgba(0, 0, 0, 0.06);
  flex-shrink: 0;
  z-index: 10;
}

.header-left {
  display: flex;
  gap: 4px;
}

.nav-btn {
  background: none;
  border: none;
  font-size: 14px;
  font-weight: 500;
  color: #6B7280;
  padding: 6px 12px;
  cursor: pointer;
  border-radius: 6px;
  transition: all 200ms ease-out;
  position: relative;
  font-family: inherit;
}

.nav-btn:hover {
  color: #4F46E5;
  background: rgba(79, 70, 229, 0.04);
}

.nav-btn::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 12px;
  right: 12px;
  height: 1.5px;
  background: #4F46E5;
  border-radius: 1px;
  transform: scaleX(0);
  transition: transform 200ms ease-out;
}

.nav-btn:hover::after {
  transform: scaleX(1);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-menu {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 200ms ease-out;
  position: relative;
  user-select: none;
}

.user-menu:hover {
  background: rgba(0, 0, 0, 0.03);
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
}

.avatar-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #4F46E5 0%, #7C3AED 100%);
  color: white;
  font-size: 13px;
  font-weight: 600;
}

.user-name {
  font-size: 14px;
  color: #374151;
  font-weight: 500;
}

.dropdown-arrow {
  font-size: 14px;
  color: #9CA3AF;
  transition: transform 200ms ease-out;
}

.dropdown-arrow.open {
  transform: rotate(180deg);
}

/* 下拉菜单 */
.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 160px;
  background: #FFFFFF;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 12px;
  box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.1), 0 1px 3px rgba(0, 0, 0, 0.04);
  padding: 6px;
  z-index: 100;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: none;
  border-radius: 8px;
  font-size: 13px;
  color: #374151;
  cursor: pointer;
  transition: all 200ms ease-out;
  font-family: inherit;
  text-align: left;
}

.dropdown-item:hover {
  background: #F3F4F6;
  color: #111827;
}

.dropdown-item.danger {
  color: #EF4444;
}

.dropdown-item.danger:hover {
  background: #FEF2F2;
  color: #DC2626;
}

.dropdown-divider {
  height: 1px;
  background: #E5E7EB;
  margin: 6px 0;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 200ms ease-out;
}

.dropdown-enter-from {
  opacity: 0;
  transform: translateY(-4px) scale(0.96);
}

.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.96);
}

/* ── 消息区域 ── */
.msg-list {
  flex: 1;
  overflow-y: auto;
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.msg-list::-webkit-scrollbar {
  width: 4px;
}

.msg-list::-webkit-scrollbar-track {
  background: transparent;
}

.msg-list::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.08);
  border-radius: 2px;
}

.msg-list::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.15);
}

.msg-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  max-width: 100%;
}

.msg-row.user {
  flex-direction: row-reverse;
}

.msg-avatar {
  flex-shrink: 0;
  margin-top: 2px;
}

.ai-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4F46E5 0%, #7C3AED 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-avatar-small {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366F1 0%, #A855F7 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 12px;
  font-weight: 600;
}

.msg-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.msg-row.ai .msg-body {
  align-items: flex-start;
}

.msg-row.user .msg-body {
  align-items: flex-end;
}

.msg-bubble {
  padding: 12px 16px;
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.msg-row.ai .msg-bubble {
  background: #FFFFFF;
  color: #1F2937;
  border: 1px solid #E5E7EB;
  border-radius: 16px;
  border-top-left-radius: 2px;
  max-width: 75%;
}

.msg-row.user .msg-bubble {
  background: #4F46E5;
  color: #FFFFFF;
  border-radius: 16px;
  border-top-right-radius: 2px;
  max-width: 70%;
  box-shadow: 0 2px 8px rgba(79, 70, 229, 0.15);
}

.msg-time {
  font-size: 11px;
  color: #9CA3AF;
  padding: 0 4px;
}

/* 消息进入动画 */
.msg-bubble-enter-active {
  transition: all 300ms ease-out;
}

.msg-bubble-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

/* ── 底部输入区 ── */
.chat-input-area {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 64px;
  padding: 0 20px;
  background: #FFFFFF;
  border-top: 1px solid #E5E7EB;
  flex-shrink: 0;
}

.attach-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #9CA3AF;
  transition: all 200ms ease-out;
  flex-shrink: 0;
}

.attach-btn:hover {
  background: #F3F4F6;
  color: #6B7280;
}

.attach-icon {
  font-size: 20px;
}

.input-wrapper {
  flex: 1;
}

.chat-input {
  width: 100%;
  height: 44px;
  border-radius: 22px;
  background: #F3F4F6;
  border: none;
  outline: none;
  padding: 0 20px;
  font-size: 14px;
  color: #1F2937;
  font-family: inherit;
  transition: all 200ms ease-out;
}

.chat-input::placeholder {
  color: #9CA3AF;
}

.chat-input:focus {
  background: #EEF2FF;
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.1);
}

.send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, #4F46E5 0%, #6366F1 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  flex-shrink: 0;
  transition: all 200ms ease-out;
}

.send-btn:hover:not(.disabled) {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.25);
}

.send-btn:active:not(.disabled) {
  transform: scale(1);
}

.send-btn.disabled {
  background: #E5E7EB;
  color: #9CA3AF;
  cursor: not-allowed;
}

.send-icon {
  font-size: 18px;
}

/* ── 响应式 ── */

/* 平板 */
@media (max-width: 1023px) {
  .chat-header {
    padding: 0 16px;
  }

  .msg-list {
    padding: 16px 16px;
  }

  .chat-input-area {
    padding: 0 16px;
  }
}

/* 手机 */
@media (max-width: 767px) {
  .chat-header {
    padding: 0 12px;
    height: 52px;
  }

  .nav-btn {
    font-size: 13px;
    padding: 5px 8px;
  }

  .user-name {
    display: none;
  }

  .user-menu {
    gap: 4px;
    padding: 4px;
  }

  .msg-list {
    padding: 12px 12px;
    gap: 16px;
  }

  .msg-row {
    gap: 8px;
  }

  .ai-avatar,
  .user-avatar-small {
    width: 28px;
    height: 28px;
  }

  .ai-avatar svg {
    width: 14px;
    height: 14px;
  }

  .msg-bubble {
    padding: 10px 13px;
    font-size: 14px;
    line-height: 1.5;
  }

  .msg-row.ai .msg-bubble {
    max-width: 85%;
  }

  .msg-row.user .msg-bubble {
    max-width: 85%;
  }

  .msg-time {
    font-size: 10px;
  }

  .chat-input-area {
    padding: 0 12px;
    height: 60px;
    gap: 8px;
  }

  .attach-btn {
    width: 36px;
    height: 36px;
  }

  .attach-icon {
    font-size: 18px;
  }

  .chat-input {
    height: 40px;
    border-radius: 20px;
    padding: 0 16px;
    font-size: 14px;
  }

  .send-btn {
    width: 36px;
    height: 36px;
  }

  .send-icon {
    font-size: 16px;
  }

  .dropdown-menu {
    right: -4px;
    width: 150px;
  }

  .dropdown-item {
    font-size: 13px;
    padding: 8px 10px;
  }
}

/* 小手机 */
@media (max-width: 375px) {
  .chat-header {
    padding: 0 10px;
  }

  .nav-btn {
    font-size: 12px;
    padding: 4px 6px;
  }

  .msg-list {
    padding: 10px 10px;
  }

  .msg-bubble {
    padding: 9px 12px;
    font-size: 13px;
  }

  .chat-input-area {
    padding: 0 10px;
    gap: 6px;
  }

  .chat-input {
    padding: 0 14px;
    font-size: 13px;
  }
}

/* 大屏限制内容宽度，避免阅读行过长 */
@media (min-width: 1440px) {
  .msg-list {
    padding-left: calc((100% - 1200px) / 2 + 24px);
    padding-right: calc((100% - 1200px) / 2 + 24px);
  }

  .chat-header,
  .chat-input-area {
    padding-left: calc((100% - 1200px) / 2 + 24px);
    padding-right: calc((100% - 1200px) / 2 + 24px);
  }
}
</style>
