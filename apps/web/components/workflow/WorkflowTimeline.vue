<script setup lang="ts">
import type { WorkflowStep, WorkflowStatus } from '~/types'
import { STATUS_LABELS } from '~/types'

const props = defineProps<{
  steps: WorkflowStep[]
  currentStatus: WorkflowStatus
  progress: number
}>()

const orderedStepNames = [
  'prepare_transcript',
  'chunk_transcript',
  'generate_summary',
  'extract_action_cards',
  'extract_claims',
  'rtrvr_research',
  'finalize',
]

const stepLabelMap: Record<string, string> = {
  prepare_transcript: 'Prepare Transcript',
  chunk_transcript: 'Chunk Content',
  generate_summary: 'Generate Summary',
  extract_action_cards: 'Extract Actions',
  extract_claims: 'Extract Claims',
  rtrvr_research: 'Browser Research',
  finalize: 'Finalize',
}

const displaySteps = computed(() => {
  const stepMap = new Map(props.steps.map(s => [s.stepName, s]))
  return orderedStepNames.map(name => ({
    name,
    label: stepLabelMap[name] ?? name.replace(/_/g, ' '),
    step: stepMap.get(name),
  }))
})

const currentStatusClass = computed(() => {
  const s = props.currentStatus
  if (s === 'ready') return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
  if (s === 'failed') return 'bg-red-500/10 text-red-400 border-red-500/20'
  if (s === 'waiting_for_user_input') return 'bg-violet-500/10 text-violet-400 border-violet-500/20'
  return 'bg-blue-500/10 text-blue-400 border-blue-500/20'
})

const currentDotClass = computed(() => {
  const s = props.currentStatus
  if (s === 'ready') return 'bg-emerald-500'
  if (s === 'failed') return 'bg-red-500'
  if (s === 'waiting_for_user_input') return 'bg-violet-500'
  return 'bg-blue-500 animate-pulse'
})

function statusColor(status?: string): string {
  switch (status) {
    case 'completed': return 'text-emerald-400'
    case 'running': return 'text-blue-400'
    case 'failed': return 'text-red-400'
    default: return 'text-zinc-500'
  }
}

function dotColor(status?: string): string {
  switch (status) {
    case 'completed': return 'bg-emerald-500'
    case 'running': return 'bg-blue-500 animate-pulse'
    case 'failed': return 'bg-red-500'
    default: return 'bg-zinc-700'
  }
}
</script>

<template>
  <div class="card p-5">
    <div class="flex items-center justify-between mb-4">
      <h3 class="font-semibold text-zinc-200 text-sm">Workflow Progress</h3>
      <span class="text-xs text-zinc-400">{{ progress }}%</span>
    </div>

    <div class="w-full bg-surface-600 rounded-full h-1.5 mb-5">
      <div
        class="bg-gradient-to-r from-accent-500 to-accent-400 h-1.5 rounded-full transition-all duration-700"
        :style="{ width: `${progress}%` }"
      />
    </div>

    <div class="mb-5">
      <span :class="['badge border text-xs', currentStatusClass]">
        <span :class="['status-dot', currentDotClass]" />
        {{ STATUS_LABELS[currentStatus] ?? currentStatus }}
      </span>
    </div>

    <ol class="space-y-3">
      <li v-for="{ name, label, step } in displaySteps" :key="name" class="flex items-center gap-3">
        <div :class="['status-dot flex-shrink-0', dotColor(step?.status)]" />
        <span :class="['text-xs flex-1', statusColor(step?.status)]">{{ label }}</span>
        <span v-if="step?.status === 'running'" class="text-xs text-blue-400 animate-pulse">running</span>
        <svg v-else-if="step?.status === 'completed'" class="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
        </svg>
        <svg v-else-if="step?.status === 'failed'" class="w-3.5 h-3.5 text-red-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </li>
    </ol>
  </div>
</template>
