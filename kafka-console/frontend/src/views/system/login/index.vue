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
