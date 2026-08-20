<script setup lang="ts">import { ref } from 'vue'; const format=ref('zip'); const expires=ref('');</script>
<template><section><h1>交付中心</h1><label>格式<select v-model="format"><option>zip</option><option>json</option><option>csv</option></select></label><label>有效期<input v-model="expires" type="datetime-local"/></label><button>提交交付申请</button><p>只有带版权授权的素材才允许导出。</p></section></template>
