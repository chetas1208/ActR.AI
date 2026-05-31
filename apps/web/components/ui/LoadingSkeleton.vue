<script setup lang="ts">
const props = defineProps<{
  lines?: number
  height?: string
}>()
const lineCount = computed(() => props.lines ?? 3)
</script>

<template>
  <div class="animate-pulse space-y-3">
    <div
      v-for="i in lineCount"
      :key="i"
      :class="['bg-surface-600 rounded-lg', i === lineCount ? 'w-2/3' : 'w-full']"
      :style="{ height: height ?? '16px' }"
    />
  </div>
</template>
