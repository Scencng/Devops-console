<template>
  <div class="login-container">
    <div class="login-content">
      <div class="login-header">
        <h2 class="title">DevOps 运维控制台</h2>
        <p class="subtitle">安全访问门户</p>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            prefix-icon="User"
            size="large"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            show-password
            size="large"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            class="login-button"
            @click="handleLogin"
            size="large"
          >
            {{ loading ? '登录中...' : '登录' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <p>&copy; 2026 peppa-pig. 版权所有.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '@/api/system/user.js'
import { SHA256 } from 'crypto-js'

const router = useRouter()
const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const loginData = {
          ...loginForm,
          password: SHA256(loginForm.password).toString()
        }
        const res = await login(loginData)
        console.log('Login Result:', res)
        // Extract token with support for nested structures
        // The user confirmed the structure: res.data.data.accessToken
        let accessToken = res.accessToken
        let refreshToken = res.refreshToken

        // Check res.data
        if (res.data) {
          // If the token is directly in res.data
          accessToken = accessToken || res.data.accessToken || res.data.access_token
          refreshToken = refreshToken || res.data.refreshToken || res.data.refresh_token

          // Check res.data.data (Double nested as seen in user logs)
          if (res.data.data) {
             accessToken = accessToken || res.data.data.accessToken || res.data.data.access_token
             refreshToken = refreshToken || res.data.data.refreshToken || res.data.data.refresh_token
          }
        }

        console.log('Final extracted AccessToken:', accessToken)

        if (accessToken) {
          localStorage.setItem('access_token', accessToken)
          console.log('Token successfully saved to localStorage. Verification:', localStorage.getItem('access_token'))
          
          if (refreshToken) {
            localStorage.setItem('refresh_token', refreshToken)
          }
          ElMessage.success('登录成功')
          router.push('/')
        } else {
          console.error('FAILED to extract token. Response nested structure check failed.', res)
          ElMessage.error('登录失败：无法获取访问令牌')
          // Clean up
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
        }
      } catch (error) {
        ElMessage.error(error.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #1e1e2e 0%, #2d2d44 100%);
  position: relative;
  width: 100%;
  overflow: hidden;
}

/* Glassmorphism Effect */
.login-content {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 40px 60px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  gap: 30px;
  z-index: 10;
}

.login-header {
  text-align: center;
}

.title {
  font-size: 2rem;
  font-weight: 700;
  color: #fff;
  margin: 0;
  background: linear-gradient(45deg, #40c9ff, #e81cff);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 10px;
}

.subtitle {
  color: #a0a0b0;
  margin: 0;
  font-size: 1rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

:deep(.el-input__wrapper) {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: none;
  border-radius: 8px;
}

:deep(.el-input__wrapper:hover), :deep(.el-input__wrapper.is-focus) {
  border-color: #40c9ff;
  background: rgba(0, 0, 0, 0.3);
}

:deep(.el-input__inner) {
  color: #fff;
}

.login-button {
  width: 100%;
  background: linear-gradient(45deg, #40c9ff, #e81cff);
  border: none;
  font-weight: bold;
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
  transition: opacity 0.3s ease;
}

.login-button:hover {
  opacity: 0.9;
}

.login-footer {
  text-align: center;
  color: #666;
  font-size: 12px;
}

/* Background decorations */
.login-container::before {
  content: '';
  position: absolute;
  top: -100px;
  left: -100px;
  width: 500px;
  height: 500px;
  background: radial-gradient(circle, #40c9ff 0%, transparent 70%);
  opacity: 0.1;
  border-radius: 50%;
}

.login-container::after {
  content: '';
  position: absolute;
  bottom: -150px;
  right: -50px;
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, #e81cff 0%, transparent 70%);
  opacity: 0.1;
  border-radius: 50%;
}
</style>
