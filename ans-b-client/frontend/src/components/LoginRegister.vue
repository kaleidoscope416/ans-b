<script setup>
import { ref, reactive, computed, watch } from 'vue'
import {
  UserIcon,
  AssignmentUserIcon,
  LockOnIcon,
  BrowseIcon,
  BrowseOffIcon,
  SecuredIcon,
} from 'tdesign-icons-vue-next'

const emit = defineEmits(['login-success'])

const activeTab = ref('login')

const loginForm = reactive({
  account: '',
  password: '',
})
const loginShowPwd = ref(false)

const registerForm = reactive({
  name: '',
  account: '',
  password: '',
  confirmPassword: '',
})
const registerShowPwd = ref(false)
const registerShowConfirm = ref(false)

const passwordStrength = computed(() => {
  const pwd = registerForm.password
  if (!pwd) return 0
  let score = 0
  if (pwd.length >= 6) score++
  if (pwd.length >= 10) score++
  if (/[A-Z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd)) score++
  if (/[^A-Za-z0-9]/.test(pwd)) score++
  if (score <= 2) return 1
  if (score <= 4) return 2
  return 3
})

const strengthColor = computed(() => {
  const map = { 1: '#EF4444', 2: '#F59E0B', 3: '#10B981' }
  return map[passwordStrength.value] || '#E5E7EB'
})

const strengthText = computed(() => {
  const map = { 1: '弱', 2: '中', 3: '强' }
  return map[passwordStrength.value] || ''
})

const strengthWidth = computed(() => {
  const map = { 1: '33%', 2: '66%', 3: '100%' }
  return map[passwordStrength.value] || '0%'
})

// TODO: 登录验证 —— 当前为 Mock 逻辑，直接通过
// TODO: 接入后端 API 进行真实登录鉴权
// TODO: 添加表单校验（帐号格式、密码强度）
// TODO: 添加记住密码、验证码等功能
function handleLogin() {
  // Mock: 模拟登录成功，直接跳转问答界面
  const name = loginForm.account.trim() || '用户'
  emit('login-success', name)
}

// TODO: 注册验证 —— 当前为 Mock 逻辑
// TODO: 接入后端 API 进行真实注册
// TODO: 添加表单校验、验证码、重复帐号检测
function handleRegister() {
  // Mock: 模拟注册成功后自动登录
  const name = registerForm.name.trim() || registerForm.account.trim() || '用户'
  emit('login-success', name)
}
</script>

<template>
  <div class="auth-card">
    <div class="glass-bg"></div>
    <div class="card-inner">
    <!-- 品牌头部 -->
    <div class="brand-header">
      <div class="brand-logo">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="10" fill="url(#logoGrad)" />
          <path d="M12 20L19 13L26 20L19 27L12 20Z" fill="white" fill-opacity="0.95" />
          <circle cx="26" cy="14" r="4" fill="white" fill-opacity="0.6" />
          <defs>
            <linearGradient id="logoGrad" x1="0" y1="0" x2="40" y2="40">
              <stop offset="0%" stop-color="#4F46E5" />
              <stop offset="100%" stop-color="#7C3AED" />
            </linearGradient>
          </defs>
        </svg>
      </div>
      <h1 class="brand-title">智问答</h1>
      <p class="brand-subtitle">开始你的智能问答之旅</p>
    </div>

    <!-- 选项卡 -->
    <div class="tabs-bar">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'login' }"
        @click="activeTab = 'login'"
      >
        登录
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'register' }"
        @click="activeTab = 'register'"
      >
        注册
      </button>
    </div>

    <Transition name="panel-fade" mode="out-in">
    <!-- 登录表单 -->
    <div v-if="activeTab === 'login'" class="form-panel" key="login">
      <div class="form-field">
        <div class="input-wrap">
          <UserIcon class="field-icon" />
          <input
            v-model="loginForm.account"
            type="text"
            placeholder="帐号"
            class="custom-input"
          />
        </div>
      </div>

      <div class="form-field">
        <div class="input-wrap">
          <LockOnIcon class="field-icon" />
          <input
            v-model="loginForm.password"
            :type="loginShowPwd ? 'text' : 'password'"
            placeholder="密码"
            class="custom-input"
          />
          <button class="eye-btn" @click="loginShowPwd = !loginShowPwd">
            <BrowseIcon v-if="loginShowPwd" class="eye-icon" />
            <BrowseOffIcon v-else class="eye-icon" />
          </button>
        </div>
      </div>

      <button class="submit-btn" @click="handleLogin">登录</button>

      <div class="form-footer">
        <a class="text-link">忘记密码？</a>
      </div>
    </div>

    <!-- 注册表单 -->
    <div v-else class="form-panel" key="register">
      <div class="form-field">
        <div class="input-wrap">
          <AssignmentUserIcon class="field-icon" />
          <input
            v-model="registerForm.name"
            type="text"
            placeholder="名称"
            class="custom-input"
          />
        </div>
      </div>

      <div class="form-field">
        <div class="input-wrap">
          <UserIcon class="field-icon" />
          <input
            v-model="registerForm.account"
            type="text"
            placeholder="帐号"
            class="custom-input"
          />
        </div>
      </div>

      <div class="form-field">
        <div class="input-wrap">
          <LockOnIcon class="field-icon" />
          <input
            v-model="registerForm.password"
            :type="registerShowPwd ? 'text' : 'password'"
            placeholder="密码"
            class="custom-input"
          />
          <button class="eye-btn" @click="registerShowPwd = !registerShowPwd">
            <BrowseIcon v-if="registerShowPwd" class="eye-icon" />
            <BrowseOffIcon v-else class="eye-icon" />
          </button>
        </div>
      </div>

      <!-- 密码强度条 -->
      <div v-if="registerForm.password" class="strength-bar">
        <div class="strength-track">
          <div
            class="strength-fill"
            :style="{ width: strengthWidth, backgroundColor: strengthColor }"
          />
        </div>
        <span class="strength-text" :style="{ color: strengthColor }">
          {{ strengthText }}
        </span>
      </div>

      <div class="form-field">
        <div class="input-wrap">
          <SecuredIcon class="field-icon" />
          <input
            v-model="registerForm.confirmPassword"
            :type="registerShowConfirm ? 'text' : 'password'"
            placeholder="确认密码"
            class="custom-input"
          />
          <button class="eye-btn" @click="registerShowConfirm = !registerShowConfirm">
            <BrowseIcon v-if="registerShowConfirm" class="eye-icon" />
            <BrowseOffIcon v-else class="eye-icon" />
          </button>
        </div>
      </div>

      <button class="submit-btn" @click="handleRegister">注册</button>
    </div>
    </Transition>

    <!-- 底部协议 -->
    <p class="agreement-text">
      登录即代表同意<a class="inline-link">用户协议</a>和<a class="inline-link">隐私政策</a>
    </p>
    </div>
  </div>
</template>

<style scoped>
.auth-card {
  position: relative;
  width: 420px;
  border-radius: 16px;
}

.glass-bg {
  position: absolute;
  inset: 0;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.45);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.7);
  box-shadow: 0 10px 40px -10px rgba(79, 70, 229, 0.06), 0 1px 3px rgba(0, 0, 0, 0.02);
  z-index: 0;
}

.card-inner {
  position: relative;
  z-index: 1;
  padding: 40px 36px 32px;
}

.brand-header {
  text-align: center;
  margin-bottom: 28px;
}

.brand-logo {
  display: inline-flex;
  margin-bottom: 14px;
}

.brand-title {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 600;
  color: #111827;
  letter-spacing: 1px;
}

.brand-subtitle {
  margin: 0;
  font-size: 14px;
  color: #6B7280;
  font-weight: 400;
}

.tabs-bar {
  display: flex;
  justify-content: center;
  gap: 32px;
  margin-bottom: 24px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  padding-bottom: 0;
}

.tab-btn {
  position: relative;
  background: none;
  border: none;
  font-size: 15px;
  font-weight: 500;
  color: #9CA3AF;
  padding: 8px 4px;
  cursor: pointer;
  transition: color 200ms ease-out;
}

.tab-btn:hover {
  color: #6B7280;
}

.tab-btn.active {
  color: #111827;
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 0;
  right: 0;
  height: 2px;
  background: #4F46E5;
  border-radius: 2px 2px 0 0;
  transition: all 200ms ease-out;
}

.panel-fade-enter-active,
.panel-fade-leave-active {
  transition: all 100ms ease-out;
}

.panel-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.panel-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.form-field {
  margin-bottom: 14px;
}

.input-wrap {
  display: flex;
  align-items: center;
  height: 44px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(0, 0, 0, 0.04);
  padding: 0 12px;
  gap: 8px;
  transition: all 200ms ease-out;
  box-sizing: border-box;
}

.input-wrap:focus-within {
  border-color: #4F46E5;
  box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.1);
  background: rgba(255, 255, 255, 0.85);
}

.field-icon {
  flex-shrink: 0;
  color: #9CA3AF;
  font-size: 18px;
  transition: color 200ms ease-out;
}

.input-wrap:focus-within .field-icon {
  color: #4F46E5;
}

.custom-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: #111827;
  height: 100%;
  padding: 0;
  font-family: inherit;
}

.custom-input::placeholder {
  color: #9CA3AF;
}

.eye-btn {
  background: none;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  color: #9CA3AF;
  transition: color 200ms ease-out;
}

.eye-btn:hover {
  color: #6B7280;
}

.eye-icon {
  font-size: 18px;
}

.strength-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
  margin-top: -4px;
}

.strength-track {
  flex: 1;
  height: 4px;
  background: #E5E7EB;
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 2px;
  transition: all 200ms ease-out;
}

.strength-text {
  font-size: 12px;
  font-weight: 500;
  min-width: 1.5em;
  text-align: right;
  transition: color 200ms ease-out;
}

.submit-btn {
  width: 100%;
  height: 44px;
  border-radius: 10px;
  border: none;
  background: linear-gradient(135deg, #4F46E5 0%, #6366F1 100%);
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  margin-top: 4px;
  font-family: inherit;
  transition: all 200ms ease-out;
}

.submit-btn:hover {
  background: linear-gradient(135deg, #5B21B6 0%, #7C3AED 100%);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(91, 33, 182, 0.2);
}

.submit-btn:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(91, 33, 182, 0.15);
}

.form-footer {
  text-align: center;
  margin-top: 16px;
}

.text-link {
  font-size: 13px;
  color: #6B7280;
  text-decoration: none;
  cursor: pointer;
  transition: color 200ms ease-out;
}

.text-link:hover {
  color: #4F46E5;
}

.agreement-text {
  text-align: center;
  font-size: 12px;
  color: #9CA3AF;
  margin: 20px 0 0;
  line-height: 1.5;
}

.inline-link {
  color: #4F46E5;
  text-decoration: none;
  cursor: pointer;
  transition: opacity 200ms ease-out;
}

.inline-link:hover {
  opacity: 0.8;
  text-decoration: underline;
}

/* 响应式 */
@media (max-width: 480px) {
  .auth-card {
    width: 90vw;
    padding: 28px 22px 24px;
  }
}
</style>
