import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useSessionStore } from '@/stores/session'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true, title: '登录' },
  },
  {
    path: '/',
    component: () => import('@/views/AppLayout.vue'),
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { title: '工作台' },
      },
      {
        path: 'projects',
        name: 'projects',
        component: () => import('@/views/ProjectsView.vue'),
        meta: { title: '项目' },
      },
      {
        path: 'cycles/:id',
        name: 'cycle-detail',
        component: () => import('@/views/CycleDetailView.vue'),
        meta: { title: '结算周期详情' },
      },
      {
        path: 'exceptions',
        name: 'exceptions',
        component: () => import('@/views/ExceptionsView.vue'),
        meta: { title: '异常处置' },
      },
      {
        path: 'exceptions/:id',
        name: 'exception-detail',
        component: () => import('@/views/ExceptionDetailView.vue'),
        meta: { title: '异常详情' },
      },
      {
        path: 'summaries',
        name: 'summaries',
        component: () => import('@/views/SummariesView.vue'),
        meta: { title: '汇总与重算' },
      },
      {
        path: 'audit',
        name: 'audit',
        component: () => import('@/views/AuditView.vue'),
        meta: { title: '审计与导出' },
      },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: { name: 'dashboard' } },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const session = useSessionStore()
  if (!to.meta.public && !session.isAuthenticated) {
    return { name: 'login' }
  }
  if (to.name === 'login' && session.isAuthenticated) {
    return { name: 'dashboard' }
  }
  return true
})

router.afterEach((to) => {
  const title = (to.meta.title as string) || ''
  document.title = title ? `${title} | 公益结算异常处置平台` : '公益结算异常处置平台'
})

export default router
