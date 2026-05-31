<script setup lang="ts">
import type { ActionCard } from '~/types'

const props = defineProps<{ card: ActionCard }>()
const emit = defineEmits<{ run: [cardId: string] }>()

const api = useApi()
const isRunning = ref(false)
const runError = ref<string | null>(null)

const typeIcon = computed(() => {
  const icons: Record<string, string> = {
    code: '⟨/⟩',
    research: '🔍',
    checklist: '✓',
    study_plan: '📚',
    browser_action: '🌐',
  }
  return icons[props.card.actionType] ?? '▶'
})

const statusClass = computed(() => {
  const map: Record<string, string> = {
    pending: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    running: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
    completed: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    failed: 'bg-red-500/10 text-red-400 border-red-500/20',
  }
  return map[props.card.status] ?? 'bg-surface-600 text-zinc-400'
})

const providerColor = computed(() => {
  const map: Record<string, string> = {
    daytona: 'text-violet-400',
    rtrvr: 'text-pink-400',
    none: 'text-zinc-500',
  }
  return map[props.card.provider ?? 'none'] ?? 'text-zinc-500'
})

async function runCard() {
  if (isRunning.value) return
  isRunning.value = true
  runError.value = null
  try {
    await api.runAction(props.card.id)
    emit('run', props.card.id)
  } catch (e: any) {
    runError.value = e.message
  } finally {
    isRunning.value = false
  }
}
</script>

<template>
  <div class="card p-4 hover:border-zinc-700/80 transition-colors">
    <div class="flex items-start gap-3">
      <!-- Type icon -->
      <div class="w-8 h-8 rounded-lg bg-surface-600 flex items-center justify-center flex-shrink-0 text-sm font-mono">
        {{ typeIcon }}
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <span class="font-medium text-zinc-200 text-sm truncate">{{ card.title }}</span>
          <span :class="['badge border text-xs flex-shrink-0', statusClass]">{{ card.status }}</span>
        </div>
        <p class="text-zinc-400 text-xs leading-relaxed mb-2">{{ card.description }}</p>

        <div class="flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <span v-if="card.provider && card.provider !== 'none'" :class="['text-xs font-medium', providerColor]">
              {{ card.provider }}
            </span>
            <span v-if="card.timestampSeconds" class="text-xs text-zinc-500">
              {{ formatTime(card.timestampSeconds) }}
            </span>
          </div>

          <button
            v-if="card.requiresExecution && card.status === 'pending'"
            @click="runCard"
            :disabled="isRunning"
            class="btn-secondary text-xs py-1.5 px-3"
          >
            <span v-if="isRunning" class="w-3 h-3 border border-current border-t-transparent rounded-full animate-spin" />
            {{ isRunning ? 'Running…' : 'Run' }}
          </button>
        </div>

        <p v-if="runError" class="text-red-400 text-xs mt-2">{{ runError }}</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${String(s).padStart(2, '0')}`
}
export default { methods: { formatTime } }
</script>
