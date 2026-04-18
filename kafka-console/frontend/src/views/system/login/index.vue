<template>
  <div class="login-container">
    <div class="login-content">
      <div class="login-header">
        <h2 class="title">Kafka Console</h2>
        <p class="subtitle">Kafka 控制台安全访问入口</p>
      </div>

      <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" class="login-form" @keyup.enter="handleLogin">
        <el-form-item prop="username"><el-input v-model="loginForm.username" placeholder="用户名" prefix-icon="User" size="large" /></el-form-item>
        <el-form-item prop="password"><el-input v-model="loginForm.password" type="password" placeholder="密码" prefix-icon="Lock" show-password size="large" /></el-form-item>
        <el-form-item><el-button type="primary" :loading="loading" class="login-button" @click="handleLogin" size="large">{{ loading ? '登录中...' : '登录' }}</el-button></el-form-item>
      </el-form>

      <div class="login-footer"><p>&copy; 2026 Kafka Console.</p></div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '@/api/system/user.js'
import { SHA256 } from 'crypto-js'

const router = useRouter()
const loginFormRef = ref(null)
const loading = ref(false)
const loginForm = reactive({ username: '', password: '' })
const loginRules = { username: [{ required: true, message: '请输入用户名', trigger: 'blur' }], password: [{ required: true, message: '请输入密码', trigger: 'blur' }] }

const handleLogin = async () => {
  if (!loginFormRef.value) return
  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const res = await login({ ...loginForm, password: SHA256(loginForm.password).toString() })
      const accessToken = res?.data?.data?.accessToken || res?.data?.accessToken || res?.accessToken
      const refreshToken = res?.data?.data?.refreshToken || res?.data?.refreshToken || res?.refreshToken
      if (!accessToken) {
        ElMessage.error('登录失败：无法获取访问令牌')
        return
      }
      localStorage.setItem('access_token', accessToken)
      if (refreshToken) localStorage.setItem('refresh_token', refreshToken)
      ElMessage.success('登录成功')
      router.push('/')
    } catch (error) {
      ElMessage.error(error.message || '登录失败')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  width: 100%;
  overflow: hidden;
  background:
    radial-gradient(circle at top left, rgba(64, 158, 255, 0.22), transparent 30%),
    radial-gradient(circle at bottom right, rgba(34, 197, 94, 0.18), transparent 28%),
    linear-gradient(135deg, #0f172a 0%, #132235 45%, #1b3654 100%);
}

.login-container::before,
.login-container::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  filter: blur(16px);
  opacity: 0.6;
}

.login-container::before {
  top: -120px;
  left: -80px;
  width: 320px;
  height: 320px;
  background: rgba(59, 130, 246, 0.3);
}

.login-container::after {
  right: -100px;
  bottom: -120px;
  width: 360px;
  height: 360px;
  background: rgba(16, 185, 129, 0.22);
}

.login-content {
  position: relative;
  z-index: 1;
  width: min(460px, calc(100vw - 32px));
  padding: 40px 36px 28px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 24px;
  background: rgba(10, 18, 30, 0.76);
  box-shadow: 0 30px 80px rgba(2, 6, 23, 0.45);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}

.login-header {
  margin-bottom: 28px;
}

.title {
  margin: 0 0 10px;
  font-size: 38px;
  line-height: 1;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #f8fafc;
}

.subtitle {
  margin: 0;
  color: rgba(226, 232, 240, 0.78);
  font-size: 15px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

:deep(.login-form .el-form-item) {
  margin-bottom: 8px;
}

:deep(.login-form .el-input__wrapper) {
  min-height: 50px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: none;
  transition: box-shadow 0.2s ease, transform 0.2s ease;
}

:deep(.login-form .el-input__wrapper:hover) {
  transform: translateY(-1px);
}

:deep(.login-form .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.18);
}

:deep(.login-form .el-input__inner) {
  color: #0f172a;
}

.login-button {
  width: 100%;
  height: 50px;
  margin-top: 8px;
  border: none;
  border-radius: 14px;
  font-size: 16px;
  font-weight: 700;
  background: linear-gradient(135deg, #409eff 0%, #22c55e 100%);
  box-shadow: 0 16px 32px rgba(64, 158, 255, 0.24);
}

.login-button:hover {
  opacity: 0.96;
  transform: translateY(-1px);
}

.login-footer {
  margin-top: 24px;
  color: rgba(226, 232, 240, 0.62);
  font-size: 13px;
  text-align: center;
}

@media (max-width: 640px) {
  .login-content {
    padding: 28px 20px 22px;
    border-radius: 20px;
  }

  .title {
    font-size: 30px;
  }
}
</style>
