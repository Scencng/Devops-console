<template>
  <el-card class="page-header-card" shadow="never">
    <div class="page-header mysql-page-header">
      <div class="page-header-copy">
        <span class="page-eyebrow">MySQL / {{ title }}</span>
        <h2>{{ title }}</h2>
        <p>{{ description }}</p>
      </div>

      <div v-if="metrics.length || $slots.actions" class="page-header-side">
        <div v-if="metrics.length" class="page-header-meta">
          <div v-for="metric in metrics" :key="metric.label" class="page-header-kpi">
            <span>{{ metric.label }}</span>
            <strong>{{ metric.value }}</strong>
          </div>
        </div>

        <div v-if="$slots.actions" class="page-header-actions">
          <slot name="actions" />
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description: string
  metrics?: Array<{
    label: string
    value: string | number
  }>
}>()
</script>

<style scoped>
.mysql-page-header {
  align-items: flex-start;
}
</style>
