<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useSessionStore } from '@/stores/session'
import { useNotificationsStore } from '@/stores/notifications'

const session = useSessionStore()
const notifications = useNotificationsStore()
const router = useRouter()
const route = useRoute()

const form = reactive({
  token: 'admin',
  tenant: 'default',
})

const error = ref('')

function submit() {
  if (!form.token.trim()) {
    error.value = '请输入演示令牌 (admin/operator/assignee/reviewer)'
    return
  }
  session.login(form.token.trim(), form.tenant.trim() || 'default')
  notifications.push('success', `已登录: ${form.token.trim()}`)
  const redirect = (route.query.redirect as string) || '/'
  router.replace(redirect)
}

function quick(t: string) {
  form.token = t
  submit()
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h2 class="text-xl" style="margin: 0 0 4px">公益结算异常处置平台</h2>
      <p class="muted" style="margin: 0 0 16px">离线演示环境</p>
      <div class="field">
        <label>演示令牌 (Bearer Token)</label>
        <input class="input" v-model="form.token" placeholder="admin/operator/assignee/reviewer" @keyup.enter="submit" />
      </div>
      <div class="field">
        <label>租户 ID</label>
        <input class="input" v-model="form.tenant" placeholder="default" @keyup.enter="submit" />
      </div>
      <div v-if="error" class="muted" style="color: var(--color-danger); margin-bottom: 12px">{{ error }}</div>
      <button class="btn btn-primary" style="width: 100%" @click="submit">登录</button>
      <div class="flex gap-2" style="margin-top: 16px">
        <button class="btn" @click="quick('admin')">admin</button>
        <button class="btn" @click="quick('operator')">operator</button>
        <button class="btn" @click="quick('assignee')">assignee</button>
        <button class="btn" @click="quick('reviewer')">reviewer</button>
      </div>
      <p class="muted text-sm" style="margin-top: 16px">
        说明: 离线模式下后端将 token 直接映射到演示账号,
        真实部署请替换为完整会话/JWT 校验.
      </p>
    </div>
  </div>
</template>
