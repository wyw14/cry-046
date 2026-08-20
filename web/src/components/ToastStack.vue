<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useNotificationsStore } from '@/stores/notifications'

const store = useNotificationsStore()
const { items } = storeToRefs(store)
</script>

<template>
  <div class="toast-stack" role="status" aria-live="polite">
    <div
      v-for="n in items"
      :key="n.id"
      :class="['toast', `toast-${n.kind}`]"
      @click="store.dismiss(n.id)"
    >
      {{ n.message }}
    </div>
  </div>
</template>
