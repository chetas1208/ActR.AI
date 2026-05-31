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
  const m: Record<string, string> = {
    beginner:     'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    intermediate: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    advanced:     'bg-red-500/10 text-red-400 border-red-500/20',
  }
  return m[props.difficulty ?? ''] ?? 'bg-surface-600 text-zinc-400'
})
</script>

<template>
  <div class="card p-5">
    <div class="flex items-center gap-2 mb-4">
      <h3 class="font-semibold text-zinc-200 text-sm">Summary</h3>
      <ProviderChip name="NVIDIA NIM" />
    </div>

    <template v-if="isLoading">
      <LoadingSkeleton :lines="4" />
    </template>
    <template v-else>
      <h4 class="font-semibold text-zinc-100 text-sm mb-2">{{ title }}</h4>
      <p class="text-zinc-400 text-xs leading-relaxed mb-4">{{ summary }}</p>

      <div v-if="audience || difficulty" class="flex flex-wrap gap-1.5 mb-3">
        <span v-if="audience" class="badge bg-surface-600 text-zinc-300 text-xs border border-zinc-700/50">{{ audience }}</span>
        <span v-if="difficulty" :class="['badge border text-xs', difficultyClass]">{{ difficulty }}</span>
      </div>

      <div v-if="keyTopics?.length" class="flex flex-wrap gap-1.5">
        <span
          v-for="topic in keyTopics"
          :key="topic"
          class="badge bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 text-xs"
        >{{ topic }}</span>
      </div>
    </template>
  </div>
</template>
