<script setup>
import { ref, computed } from 'vue'
import LoginRegister from './components/LoginRegister.vue'
import Chat from './components/Chat.vue'

const currentView = ref('login')
const userName = ref('用户')

const bgClass = computed(() =>
  currentView.value === 'login' ? 'bg-login' : 'bg-chat'
)

function onLoginSuccess(name) {
  userName.value = name || '用户'
  currentView.value = 'chat'
}

function onLogout() {
  currentView.value = 'login'
}
</script>

<template>
  <div class="app-root" :class="bgClass">
    <div v-if="currentView === 'login'" class="login-bg-glow" />
    <LoginRegister
      v-if="currentView === 'login'"
      @login-success="onLoginSuccess"
    />
    <Chat
      v-else
      class="fade-in"
      :user-name="userName"
      @logout="onLogout"
    />
  </div>
</template>

<style scoped>
.app-root {
  min-height: 100vh;
  display: flex;
  position: relative;
  overflow: hidden;
}

.app-root.bg-login {
  align-items: center;
  justify-content: center;
  background: radial-gradient(circle at 80% 20%, #EEF2FF 0%, #F8FAFC 50%, #FFFFFF 100%);
}

.app-root.bg-chat {
  align-items: stretch;
  justify-content: stretch;
  background: #F8FAFC;
}

.login-bg-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 800px;
  height: 800px;
  transform: translate(-50%, -50%);
  background: radial-gradient(
    circle at 50% 50%,
    rgba(148, 163, 184, 0.08) 0%,
    transparent 60%
  );
  pointer-events: none;
  z-index: 0;
}

.fade-in {
  animation: viewFadeIn 120ms ease-out;
}

@keyframes viewFadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
