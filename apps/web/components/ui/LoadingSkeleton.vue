<script setup lang="ts">
const props = defineProps<{ lines?: number; height?: string }>()
const n = computed(() => props.lines ?? 3)
</script>

<template>
  <div class="animate-pulse space-y-3">
    <div
      v-for="i in n"
      :key="i"
      :class="['bg-surface-600 rounded-lg', i === n ? 'w-2/3' : 'w-full']"
      :style="{ height: props.height ?? '14px' }"
    />
  </div>
</template>
