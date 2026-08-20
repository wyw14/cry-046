<script setup lang="ts">
import { ref } from 'vue'
import ProjectGallery from './pages/ProjectGallery.vue'
import DeliveryCenter from './pages/DeliveryCenter.vue'
import ReviewQueue from './pages/ReviewQueue.vue'
const page = ref('projects')
</script>
<template>
  <main><nav><button @click="page='projects'">项目库</button><button @click="page='review'">审核队列</button><button @click="page='delivery'">交付中心</button></nav><ProjectGallery v-if="page==='projects'"/><ReviewQueue v-else-if="page==='review'"/><DeliveryCenter v-else/></main>
</template>
