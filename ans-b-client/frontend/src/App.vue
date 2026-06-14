<script setup>
import { computed, ref } from 'vue'
import LoginRegister from './components/LoginRegister.vue'
import StudentShell from './components/StudentShell.vue'

const currentView = ref('login')
const currentUser = ref(null)

const bgClass = computed(() => (
  currentView.value === 'login' ? 'bg-login' : 'bg-shell'
))

function onLoginSuccess(user) {
  currentUser.value = user || null
  currentView.value = 'shell'
}

function onLogout() {
  currentUser.value = null
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
    <StudentShell
      v-else
      :initial-user="currentUser"
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
  background:
    radial-gradient(circle at top right, rgba(249, 115, 22, 0.10), transparent 28%),
    linear-gradient(135deg, #fef7ed 0%, #fff7ed 20%, #eff6ff 72%, #ffffff 100%);
}

.app-root.bg-shell {
  background:
    radial-gradient(circle at top left, rgba(14, 165, 233, 0.08), transparent 24%),
    linear-gradient(180deg, #fffdf8 0%, #f7fbff 48%, #f4f7fb 100%);
}

.login-bg-glow {
  position: absolute;
  inset: auto auto 10% 10%;
  width: 720px;
  height: 720px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(14, 165, 233, 0.08) 0%, transparent 62%);
  pointer-events: none;
}
</style>
