<template>
  <div class="login-page">
    <div class="scene scene-sky" aria-hidden="true"></div>
    <div class="scene scene-clouds" aria-hidden="true"></div>
    <div class="scene scene-ridge scene-ridge--back" aria-hidden="true"></div>
    <div class="scene scene-ridge scene-ridge--mid" aria-hidden="true"></div>
    <div class="scene scene-ridge scene-ridge--front" aria-hidden="true"></div>
    <div class="scene scene-vignette" aria-hidden="true"></div>

    <main class="login-shell">
      <section class="login-panel fade-up" role="main">
        <div class="panel-header">
          <p class="panel-kicker">{{ LOGIN_COPY.kicker }}</p>
          <h1 class="panel-title">{{ LOGIN_COPY.title }}</h1>
          <p class="panel-subtitle">{{ LOGIN_COPY.subtitle }}</p>
        </div>

        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username" class="line-field">
            <label class="field-label" for="login-username">{{ LOGIN_COPY.usernameLabel }}</label>
            <el-input
              ref="usernameInputRef"
              v-model="loginForm.username"
              input-id="login-username"
              class="field-input"
              :placeholder="LOGIN_COPY.usernamePlaceholder"
              autocomplete="username"
            />
          </el-form-item>

          <el-form-item prop="password" class="line-field">
            <label class="field-label" for="login-password">{{ LOGIN_COPY.passwordLabel }}</label>
            <el-input
              v-model="loginForm.password"
              input-id="login-password"
              class="field-input"
              type="password"
              :placeholder="LOGIN_COPY.passwordPlaceholder"
              show-password
              autocomplete="current-password"
            />
          </el-form-item>

          <div class="login-meta">
            <label class="remember-wrap">
              <input v-model="rememberMe" class="remember-input" type="checkbox" />
              <span>{{ LOGIN_COPY.rememberMe }}</span>
            </label>
            <button type="button" class="ghost-link" @click="showPendingMessage(LOGIN_COPY.forgotPasswordFeature)">
              {{ LOGIN_COPY.forgotPassword }}
            </button>
          </div>

          <el-form-item class="submit-row">
            <el-button
              type="primary"
              :loading="loading"
              class="submit-btn"
              @click="handleLogin"
            >
              {{ loading ? LOGIN_COPY.loginLoading : LOGIN_COPY.loginAction }}
            </el-button>
          </el-form-item>
        </el-form>

        <p class="panel-footer">
          {{ LOGIN_COPY.footerPrefix }}
          <button type="button" class="inline-link" @click="showPendingMessage(LOGIN_COPY.registerFeature)">
            {{ LOGIN_COPY.footerAction }}
          </button>
        </p>
      </section>
    </main>
  </div>
</template>

<script setup>
import { nextTick, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { SHA256 } from 'crypto-js'

import { login } from '@/api/system/user.js'
import {
  buildPendingMessage,
  extractTokens,
  loadRememberedUsername,
  LOGIN_COPY,
  persistRememberedUsername,
  persistLoginTokens,
  resolveLoginRedirect
} from '@/utils/loginHelpers.js'

const router = useRouter()
const route = useRoute()

const loginFormRef = ref(null)
const usernameInputRef = ref(null)
const loading = ref(false)
const rememberMe = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [{ required: true, message: LOGIN_COPY.validationUsername, trigger: 'blur' }],
  password: [{ required: true, message: LOGIN_COPY.validationPassword, trigger: 'blur' }]
}

const focusUsername = () => {
  usernameInputRef.value?.focus?.()
}

const showPendingMessage = (featureName) => {
  ElMessage.info(buildPendingMessage(featureName, LOGIN_COPY.featurePendingSuffix))
}

const handleLogin = async () => {
  if (!loginFormRef.value || loading.value) return

  try {
    const valid = await loginFormRef.value.validate().catch(() => false)
    if (!valid) return

    loading.value = true
    const res = await login({
      ...loginForm,
      username: loginForm.username.trim(),
      password: SHA256(loginForm.password).toString()
    })

    const { accessToken, refreshToken } = extractTokens(res)

    if (!accessToken) {
      ElMessage.error(LOGIN_COPY.missingToken)
      return
    }

    persistLoginTokens({ accessToken, refreshToken })
    persistRememberedUsername({
      rememberMe: rememberMe.value,
      username: loginForm.username
    })

    ElMessage.success(LOGIN_COPY.loginSuccess)
    router.push(resolveLoginRedirect(route))
  } catch (error) {
    ElMessage.error(error.message || LOGIN_COPY.loginFallbackError)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  const rememberedUsername = loadRememberedUsername()
  if (rememberedUsername) {
    loginForm.username = rememberedUsername
    rememberMe.value = true
  }
  await nextTick()
  focusUsername()
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;600;700;800&family=Plus+Jakarta+Sans:wght@500;600;700;800&display=swap');

.login-page {
  --sky-top-rgb: 103, 117, 140;
  --sky-bottom-rgb: 12, 17, 29;
  --surface-rgb: 16, 24, 39;
  --surface-strong-rgb: 22, 33, 54;
  --ink-rgb: 245, 247, 252;
  --text-secondary-rgb: 211, 220, 236;
  --text-muted-rgb: 154, 168, 194;
  --white-rgb: 255, 255, 255;
  --line-rgb: 234, 239, 248;
  --accent-rgb: 131, 198, 255;
  --mountain-plum-rgb: 74, 49, 74;
  --mountain-crimson-rgb: 132, 61, 82;
  position: relative;
  flex: 1 0 100%;
  align-self: stretch;
  width: 100vw;
  min-width: 100vw;
  max-width: 100vw;
  min-height: 100svh;
  overflow: hidden;
  color: rgb(var(--ink-rgb));
  font-family: var(--font-display);
  background: rgb(var(--sky-bottom-rgb));
  isolation: isolate;
}

.scene {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.scene-sky {
  background:
    radial-gradient(48rem 16rem at 50% 8%, rgba(var(--white-rgb), 0.18), transparent 60%),
    linear-gradient(180deg, rgba(var(--sky-top-rgb), 1) 0%, rgba(50, 63, 89, 1) 24%, rgba(23, 27, 44, 1) 58%, rgba(var(--sky-bottom-rgb), 1) 100%);
}

.scene-clouds {
  top: 0;
  height: 36%;
  background:
    radial-gradient(18rem 5rem at 12% 38%, rgba(var(--white-rgb), 0.16), transparent 72%),
    radial-gradient(16rem 4rem at 34% 28%, rgba(var(--white-rgb), 0.12), transparent 72%),
    radial-gradient(22rem 5rem at 62% 34%, rgba(var(--white-rgb), 0.12), transparent 74%),
    radial-gradient(18rem 4rem at 82% 24%, rgba(var(--white-rgb), 0.08), transparent 74%);
  filter: blur(10px);
  opacity: 0.72;
}

.scene-ridge {
  inset: auto 0 0;
}

.scene-ridge--back {
  height: 62%;
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0), rgba(7, 10, 18, 0.12)),
    linear-gradient(115deg, rgba(0, 0, 0, 0) 0 18%, rgba(var(--mountain-plum-rgb), 0.9) 18% 24%, rgba(0, 0, 0, 0) 24% 31%, rgba(var(--mountain-plum-rgb), 0.86) 31% 37%, rgba(0, 0, 0, 0) 37% 44%, rgba(var(--mountain-plum-rgb), 0.92) 44% 52%, rgba(0, 0, 0, 0) 52% 59%, rgba(var(--mountain-plum-rgb), 0.84) 59% 66%, rgba(0, 0, 0, 0) 66% 72%, rgba(var(--mountain-plum-rgb), 0.88) 72% 79%, rgba(0, 0, 0, 0) 79% 100%);
  clip-path: polygon(0 48%, 8% 42%, 15% 46%, 23% 34%, 31% 38%, 38% 24%, 47% 28%, 54% 20%, 63% 35%, 70% 26%, 78% 41%, 86% 30%, 93% 46%, 100% 40%, 100% 100%, 0 100%);
  opacity: 0.92;
}

.scene-ridge--mid {
  height: 54%;
  background:
    linear-gradient(180deg, rgba(var(--mountain-crimson-rgb), 0.2), rgba(4, 5, 10, 0.22)),
    linear-gradient(90deg, rgba(var(--mountain-crimson-rgb), 0.34), rgba(var(--mountain-plum-rgb), 0.42) 45%, rgba(18, 12, 22, 0.76));
  clip-path: polygon(0 44%, 10% 52%, 18% 46%, 27% 58%, 38% 41%, 48% 61%, 58% 46%, 69% 56%, 78% 42%, 87% 64%, 100% 50%, 100% 100%, 0 100%);
  filter: saturate(1.12);
}

.scene-ridge--front {
  height: 28%;
  background:
    linear-gradient(180deg, rgba(0, 0, 0, 0), rgba(2, 4, 8, 0.92)),
    linear-gradient(90deg, rgba(6, 8, 15, 0.98), rgba(10, 13, 22, 1));
  clip-path: polygon(0 52%, 14% 48%, 26% 62%, 36% 57%, 48% 69%, 63% 54%, 76% 66%, 89% 58%, 100% 70%, 100% 100%, 0 100%);
}

.scene-vignette {
  background:
    radial-gradient(circle at 50% 42%, rgba(255, 255, 255, 0.04), transparent 26%),
    linear-gradient(180deg, rgba(3, 4, 9, 0.26), rgba(1, 1, 4, 0.54));
}

.login-shell {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-height: 100svh;
  padding: 24px;
}

.login-panel {
  width: min(100%, 640px);
  padding: 42px 46px 34px;
  border-radius: 28px;
  background:
    linear-gradient(180deg, rgba(var(--surface-strong-rgb), 0.24), rgba(var(--surface-rgb), 0.52)),
    rgba(var(--surface-rgb), 0.48);
  border: 1px solid rgba(var(--white-rgb), 0.28);
  backdrop-filter: blur(20px) saturate(118%);
  -webkit-backdrop-filter: blur(20px) saturate(118%);
  box-shadow:
    0 24px 64px rgba(0, 0, 0, 0.34),
    inset 0 1px 0 rgba(var(--white-rgb), 0.14);
}

.panel-header {
  margin-bottom: 28px;
  text-align: center;
}

.panel-kicker {
  margin: 0 0 12px;
  color: rgba(var(--text-muted-rgb), 0.9);
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.panel-title {
  margin: 0;
  color: rgb(var(--ink-rgb));
  font-size: clamp(2.8rem, 2.1103rem + 1.3793vw, 3.5rem);
  font-weight: 800;
  line-height: 0.96;
  letter-spacing: -0.05em;
  text-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.panel-subtitle {
  margin: 14px auto 0;
  max-width: 28rem;
  color: rgba(var(--text-secondary-rgb), 0.88);
  font-size: 0.95rem;
  line-height: 1.72;
}

.login-form {
  display: flex;
  flex-direction: column;
}

.line-field {
  position: relative;
  margin-bottom: 22px;
}

.field-label {
  display: block;
  margin-bottom: 6px;
  color: rgba(var(--text-secondary-rgb), 0.94);
  font-size: 0.94rem;
  font-weight: 500;
}

:deep(.line-field .el-form-item__content) {
  display: block;
}

:deep(.line-field .el-input__wrapper) {
  min-height: 52px;
  padding: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none !important;
}

:deep(.line-field .el-input__inner) {
  height: 52px;
  padding: 0;
  color: rgb(var(--ink-rgb));
  font-size: 1rem;
  font-weight: 600;
  background: transparent;
  border-bottom: 2px solid rgba(var(--line-rgb), 0.98);
  font-family: var(--font-sans);
  transition: border-color 180ms ease;
}

:deep(.line-field .el-input__inner::placeholder) {
  color: rgba(var(--text-muted-rgb), 0.7);
}

:deep(.line-field .el-input__inner:focus) {
  border-bottom-color: rgba(var(--accent-rgb), 1);
}

:deep(.line-field .el-input__suffix-inner) {
  color: rgba(var(--text-secondary-rgb), 0.9);
}

:deep(.line-field .el-form-item__error) {
  color: rgba(255, 174, 174, 0.95);
  font-size: 0.76rem;
  padding-top: 6px;
}

.login-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 14px 0 30px;
}

.remember-wrap {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  color: rgba(var(--ink-rgb), 0.94);
  font-size: 0.92rem;
  cursor: pointer;
}

.remember-input {
  width: 17px;
  height: 17px;
  margin: 0;
  accent-color: rgb(var(--white-rgb));
}

.ghost-link {
  padding: 0;
  border: none;
  color: rgba(var(--ink-rgb), 0.92);
  background: transparent;
  font-size: 0.92rem;
  cursor: pointer;
}

.ghost-link:hover {
  color: rgba(var(--accent-rgb), 1);
}

.submit-row {
  margin-bottom: 0;
}

.submit-btn {
  width: 100%;
  min-height: 58px;
  border: none;
  border-radius: 8px;
  color: rgba(10, 14, 24, 0.98);
  background: rgba(var(--white-rgb), 0.96);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.22);
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  transition: transform 180ms ease, filter 180ms ease;
}

.submit-btn:hover {
  transform: translateY(-1px);
  filter: brightness(1.02);
}

:deep(.submit-btn.el-button) {
  background: rgba(var(--white-rgb), 0.96) !important;
  border: none !important;
  color: rgba(10, 14, 24, 0.98) !important;
}

.panel-footer {
  margin: 28px 0 0;
  color: rgba(var(--ink-rgb), 0.9);
  font-size: 0.92rem;
  text-align: center;
}

.inline-link {
  padding: 0;
  border: none;
  color: inherit;
  background: transparent;
  font-weight: 700;
  font-size: inherit;
  cursor: pointer;
}

.inline-link:hover {
  color: rgba(var(--accent-rgb), 1);
}

.fade-up {
  opacity: 0;
  transform: translateY(28px);
  animation: panelFadeUp 780ms cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes panelFadeUp {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .login-shell {
    padding: 16px;
  }

  .login-panel {
    padding: 32px 26px 28px;
  }
}

@media (max-width: 420px) {
  .login-panel {
    padding: 28px 20px 24px;
    border-radius: 24px;
  }

  .panel-title {
    font-size: 2.45rem;
  }

  .login-meta {
    flex-direction: column;
    align-items: flex-start;
  }
}

@media (max-width: 360px), (max-height: 680px) {
  .login-shell {
    padding: 12px;
  }

  .login-panel {
    padding: 24px 18px 22px;
  }

  .panel-title {
    font-size: 2.2rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .fade-up {
    opacity: 1;
    transform: none;
    animation: none !important;
  }

  .submit-btn,
  :deep(.line-field .el-input__inner) {
    transition: none !important;
  }
}
</style>
