<script setup lang="ts">
const props = defineProps<{
  title: string
  summary: string
  audience?: string
  difficulty?: string
  keyTopics?: string[]
  isLoading?: boolean
}>()

const difficultyClass = computed(() => {
  const map: Record<string, string> = {
    beginner: 'bg-emerald-500/10 text-emerald-400',
    intermediate: 'bg-amber-500/10 text-amber-400',
    advanced: 'bg-red-500/10 text-red-400',
  }
  return map[props.difficulty ?? ''] ?? 'bg-surface-600 text-zinc-400'
})
</script>

<template>
  <div class="card p-5">
    <h3 class="font-semibold text-zinc-200 text-sm mb-4">Summary</h3>

    <template v-if="isLoading">
      <LoadingSkeleton :lines="4" />
    </template>
    <template v-else>
      <h4 class="font-semibold text-zinc-100 mb-2">{{ title }}</h4>
      <p class="text-zinc-400 text-sm leading-relaxed mb-4">{{ summary }}</p>

      <div v-if="audience || difficulty" class="flex flex-wrap gap-2 mb-3">
        <span v-if="audience" class="badge bg-surface-600 text-zinc-300 text-xs">{{ audience }}</span>
        <span v-if="difficulty" :class="['badge text-xs', difficultyClass]">{{ difficulty }}</span>
      </div>

      <div v-if="keyTopics?.length" class="flex flex-wrap gap-1.5">
        <span
          v-for="topic in keyTopics"
          :key="topic"
          class="badge bg-accent-500/10 text-accent-300 border border-accent-500/20 text-xs"
        >{{ topic }}</span>
      </div>
    </template>
  </div>
</template>
