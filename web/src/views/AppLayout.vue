<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute, RouterView, RouterLink } from 'vue-router'
import { useSessionStore } from '@/stores/session'

const session = useSessionStore()
const router = useRouter()
const route = useRoute()

const navItems = computed(() => [
  { name: 'dashboard', label: '工作台', to: '/' },
  { name: 'projects', label: '项目', to: '/projects' },
  { name: 'exceptions', label: '异常处置', to: '/exceptions' },
  { name: 'summaries', label: '汇总与重算', to: '/summaries' },
  { name: 'audit', label: '审计与导出', to: '/audit' },
])

function logout() {
  session.logout()
  router.replace({ name: 'login' })
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <h1>公益结算异常处置</h1>
      <RouterLink
        v-for="item in navItems"
        :key="item.name"
        :to="item.to"
        active-class="router-link-active"
      >
        {{ item.label }}
      </RouterLink>
      <div class="tenant">
        <div>当前租户: {{ session.tenant }}</div>
        <div>当前用户: {{ session.user?.username || '-' }} ({{ session.user?.role || '-' }})</div>
        <button class="btn" style="margin-top: 8px; width: 100%" @click="logout">退出</button>
      </div>
    </aside>
    <main class="main-content">
      <RouterView :key="route.fullPath" />
    </main>
  </div>
</template>
