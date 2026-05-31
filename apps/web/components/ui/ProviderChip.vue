<script setup lang="ts">
defineProps<{ name: string; active?: boolean }>()
</script>

<template>
  <span :class="['provider-chip transition-colors', active ? 'border-cyan-500/40 text-cyan-400 bg-cyan-500/10' : '']">
    {{ name }}
  </span>
</template>
