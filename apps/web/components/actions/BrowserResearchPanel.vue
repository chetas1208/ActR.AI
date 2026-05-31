<script setup lang="ts">
import type { BrowserRun } from '~/types'

const props = defineProps<{
  runs: BrowserRun[]
  jobId: string
}>()

const emit = defineEmits<{ ran: [] }>()

const api = useApi()
const isRunning = ref(false)
const runError = ref<string | null>(null)
const customTask = ref('')

async function runResearch() {
  if (!customTask.value.trim() || isRunning.value) return
  isRunning.value = true
  runError.value = null
  try {
    await api.runRtrvr(props.jobId, customTask.value.trim())
    customTask.value = ''
    emit('ran')
  } catch (e: any) {
    runError.value = e.message
  } finally {
    isRunning.value = false
  }
}

function statusConfig(status: string) {
  const map: Record<string, { cls: string; dot: string }> = {
    pending:   { cls: 'bg-amber-500/10 text-amber-400 border-amber-500/20',   dot: 'bg-amber-500' },
    running:   { cls: 'bg-blue-500/10 text-blue-400 border-blue-500/20',     dot: 'bg-blue-500 animate-pulse' },
    completed: { cls: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20', dot: 'bg-emerald-500' },
    failed:    { cls: 'bg-red-500/10 text-red-400 border-red-500/20',        dot: 'bg-red-500' },
  }
  return map[status] ?? map.pending
}
</script>

<template>
  <div class="card p-5">
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <h3 class="font-semibold text-zinc-200 text-sm">Browser Research</h3>
        <SponsorBadge name="Rtrvr" color="pink" />
      </div>
    </div>

    <!-- Run history -->
    <div v-if="runs.length" class="space-y-3 mb-4">
      <div
        v-for="run in runs"
        :key="run.id"
        class="bg-surface-700 rounded-xl p-3"
      >
        <div class="flex items-center gap-2 mb-1.5">
          <span :class="['status-dot flex-shrink-0', statusConfig(run.status).dot]" />
          <span :class="['badge border text-xs', statusConfig(run.status).cls]">{{ run.status }}</span>
          <span class="text-zinc-500 text-xs">{{ run.provider }}</span>
        </div>
        <p class="text-zinc-300 text-xs leading-relaxed">{{ run.task }}</p>
        <p v-if="run.outputKey" class="text-zinc-500 text-xs mt-1.5 font-mono truncate">{{ run.outputKey }}</p>
      </div>
    </div>

    <!-- Ad-hoc research trigger -->
    <form @submit.prevent="runResearch" class="space-y-3">
      <div>
        <label class="block text-xs font-medium text-zinc-400 mb-1.5">Run browser research task</label>
        <input
          v-model="customTask"
          type="text"
          placeholder="e.g. Find official documentation for React hooks"
          class="input-base text-sm"
          :disabled="isRunning"
        />
      </div>
      <ErrorState v-if="runError" :message="runError" />
      <button
        type="submit"
        :disabled="isRunning || !customTask.trim()"
        class="btn-secondary text-xs w-full justify-center"
      >
        <span v-if="isRunning" class="w-3.5 h-3.5 border border-current border-t-transparent rounded-full animate-spin" />
        {{ isRunning ? 'Running…' : 'Start Browser Research' }}
      </button>
    </form>

    <p v-if="!runs.length && !isRunning" class="text-zinc-500 text-xs mt-3">
      No browser research runs yet. Run a task above or wait for the workflow to trigger one automatically.
    </p>
  </div>
</template>
