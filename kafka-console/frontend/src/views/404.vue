<template>
  <div class="not-found">
    <div class="content">
      <div class="code">404</div>
      <div class="message">页面不存在</div>
      <div class="desc">您访问的页面已被删除或无访问权限</div>
      <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
.not-found {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}
.content {
  text-align: center;
}
.code {
  font-size: 120px;
  font-weight: 700;
  color: #409EFF;
  line-height: 1;
}
.message {
  font-size: 28px;
  color: #303133;
  margin: 16px 0 8px;
}
.desc {
  font-size: 14px;
  color: #909399;
  margin-bottom: 24px;
}
</style>
