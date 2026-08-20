<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <div class="page-header">
    <div>
      <h2 class="text-xl" style="margin: 0">{{ title }}</h2>
      <p v-if="description" class="muted text-sm" style="margin: 4px 0 0">{{ description }}</p>
    </div>
    <slot name="actions" />
  </div>
</template>
