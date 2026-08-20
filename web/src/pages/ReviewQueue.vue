<script setup lang="ts">const states=['草稿','待审核','已批准','已交付','已撤回','已归档']</script>
<template><section><h1>审核队列</h1><div v-for="state in states" :key="state" class="lane"><h2>{{state}}</h2><p>拖入该状态的方案会在此展示</p></div></section></template>
