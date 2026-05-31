<script setup lang="ts">
const props = defineProps<{
  name: string
  color?: 'blue' | 'green' | 'purple' | 'orange' | 'pink' | 'indigo'
}>()

const colorClasses = computed(() => {
  const map: Record<string, string> = {
    blue:   'bg-blue-500/10 text-blue-400 border-blue-500/20',
    green:  'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    purple: 'bg-violet-500/10 text-violet-400 border-violet-500/20',
    orange: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
    pink:   'bg-pink-500/10 text-pink-400 border-pink-500/20',
    indigo: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
  }
  return map[props.color ?? 'indigo']
})
</script>

<template>
  <span :class="['badge border', colorClasses]">{{ name }}</span>
</template>
