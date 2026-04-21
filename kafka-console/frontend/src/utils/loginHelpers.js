export const REMEMBERED_USERNAME_KEY = 'remembered_username'

export const LOGIN_COPY = {
  kicker: 'Kafka Console',
  title: '登录',
  subtitle: '登录到 Kafka 运维控制台，继续当前集群的治理与排障工作。',
  usernameLabel: '请输入用户名',
  usernamePlaceholder: '用户名',
  passwordLabel: '请输入密码',
  passwordPlaceholder: '密码',
  rememberMe: '记住我',
  forgotPassword: '忘记密码？',
  forgotPasswordFeature: '密码找回',
  loginAction: '登录',
  loginLoading: '登录中...',
  footerPrefix: '还没有账号？',
  footerAction: '立即注册',
  registerFeature: '注册功能',
  validationUsername: '请输入用户名',
  validationPassword: '请输入密码',
  missingToken: '登录失败：无法获取访问令牌',
  loginSuccess: '登录成功',
  loginFallbackError: '登录失败',
  featurePendingSuffix: '暂未开放'
}

export const extractTokens = (response) => ({
  accessToken: response?.data?.data?.accessToken || response?.data?.accessToken || response?.accessToken,
  refreshToken: response?.data?.data?.refreshToken || response?.data?.refreshToken || response?.refreshToken
})

export const resolveLoginRedirect = (route) => {
  if (typeof route?.query?.redirect !== 'string') return '/'
  return route.query.redirect || '/'
}

export const loadRememberedUsername = () => localStorage.getItem(REMEMBERED_USERNAME_KEY) || ''

export const persistRememberedUsername = ({ rememberMe, username }) => {
  const normalizedUsername = username.trim()
  if (rememberMe && normalizedUsername) {
    localStorage.setItem(REMEMBERED_USERNAME_KEY, normalizedUsername)
    return
  }
  localStorage.removeItem(REMEMBERED_USERNAME_KEY)
}

export const persistLoginTokens = ({ accessToken, refreshToken }) => {
  localStorage.setItem('access_token', accessToken)
  if (refreshToken) {
    localStorage.setItem('refresh_token', refreshToken)
    return
  }
  localStorage.removeItem('refresh_token')
}

export const buildPendingMessage = (featureName, suffix) => `${featureName}${suffix}`
