<script setup lang="ts">
import type { Severity, ExceptionStatus } from '@/types'

defineProps<{
  severity?: Severity
  status?: ExceptionStatus
}>()
</script>

<template>
  <span v-if="severity" :class="['tag', `tag-${severity}`]">{{ severity }}</span>
  <span v-else-if="status" :class="['tag', `tag-${status}`]">{{ status }}</span>
</template>
