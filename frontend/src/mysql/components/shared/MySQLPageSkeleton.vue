<template>
  <div class="mysql-page-skeleton">
    <el-skeleton animated>
      <template #template>
        <div class="mysql-page-skeleton__header">
          <el-skeleton-item variant="text" style="width: 140px; height: 14px;" />
          <el-skeleton-item variant="h3" style="width: 220px; height: 28px;" />
          <el-skeleton-item variant="text" style="width: 72%; height: 16px;" />
          <div class="mysql-page-skeleton__metrics">
            <el-skeleton-item
              v-for="index in metrics"
              :key="`metric-${index}`"
              variant="rect"
              style="width: 180px; height: 76px;"
            />
          </div>
        </div>
      </template>
    </el-skeleton>

    <div class="mysql-page-skeleton__body" :class="`mysql-page-skeleton__body--${layout}`">
      <el-skeleton
        v-for="index in cards"
        :key="`card-${index}`"
        animated
        class="mysql-page-skeleton__card"
      >
        <template #template>
          <el-skeleton-item variant="h3" style="width: 180px; height: 22px;" />
          <div class="mysql-page-skeleton__toolbar">
            <el-skeleton-item variant="rect" style="width: 110px; height: 36px;" />
            <el-skeleton-item variant="rect" style="width: 110px; height: 36px;" />
            <el-skeleton-item variant="rect" style="width: 110px; height: 36px;" />
          </div>
          <div class="mysql-page-skeleton__content">
            <el-skeleton-item variant="rect" style="width: 100%; height: 18px;" />
            <el-skeleton-item variant="rect" style="width: 100%; height: 18px;" />
            <el-skeleton-item variant="rect" style="width: 92%; height: 18px;" />
            <el-skeleton-item variant="rect" style="width: 96%; height: 18px;" />
            <el-skeleton-item variant="rect" style="width: 88%; height: 18px;" />
            <el-skeleton-item variant="rect" style="width: 100%; height: 160px;" />
          </div>
        </template>
      </el-skeleton>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  cards?: number
  metrics?: number
  layout?: 'single' | 'split'
}>(), {
  cards: 1,
  metrics: 3,
  layout: 'single'
})
</script>

<style scoped>
.mysql-page-skeleton {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.mysql-page-skeleton__header,
.mysql-page-skeleton__card {
  padding: 24px;
  border: 1px solid var(--devops-border-light);
  border-radius: var(--devops-radius-lg);
  background: var(--devops-bg-panel);
  box-shadow: var(--devops-shadow-xs);
}

.mysql-page-skeleton__header {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.mysql-page-skeleton__metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 8px;
}

.mysql-page-skeleton__body {
  display: grid;
  gap: 20px;
}

.mysql-page-skeleton__body--single {
  grid-template-columns: 1fr;
}

.mysql-page-skeleton__body--split {
  grid-template-columns: minmax(320px, 380px) minmax(0, 1fr);
}

.mysql-page-skeleton__card {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.mysql-page-skeleton__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.mysql-page-skeleton__content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

@media (max-width: 1200px) {
  .mysql-page-skeleton__body--split {
    grid-template-columns: 1fr;
  }
}
</style>
