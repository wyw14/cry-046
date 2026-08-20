<script setup lang="ts">
import type { PageResult } from '@/types'

defineProps<{
  page: PageResult
}>()

defineEmits<{
  (e: 'change', page: number): void
}>()
</script>

<template>
  <div class="flex items-center gap-3 text-sm" style="padding: 8px 0">
    <span class="muted">共 {{ page.total }} 条</span>
    <span class="muted">· 第 {{ page.page }} / {{ Math.max(1, Math.ceil(page.total / page.page_size)) }} 页</span>
    <div class="flex gap-2">
      <button class="btn" :disabled="page.page <= 1" @click="$emit('change', page.page - 1)">上一页</button>
      <button class="btn" :disabled="!page.has_next" @click="$emit('change', page.page + 1)">下一页</button>
    </div>
  </div>
</template>
