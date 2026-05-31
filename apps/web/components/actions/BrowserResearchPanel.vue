<script setup lang="ts">
import type { BrowserRun } from '~/types'

const props = defineProps<{ runs: BrowserRun[]; jobId: string }>()
const emit  = defineEmits<{ ran: [] }>()

const api        = useApi()
const isRunning  = ref(false)
const runError   = ref<string | null>(null)
const customTask = ref('')

async function runResearch() {
  if (!customTask.value.trim() || isRunning.value) return
  isRunning.value = true; runError.value = null
  try {
    await api.runRtrvr(props.jobId, customTask.value.trim())
    customTask.value = ''
    emit('ran')
  } catch (e: any) { runError.value = e.message }
  finally { isRunning.value = false }
}

function statusDot(s: string): string {
  return { pending: 'bg-amber-500', running: 'bg-cyan-500 animate-pulse', completed: 'bg-emerald-500', failed: 'bg-red-500' }[s] ?? 'bg-zinc-500'
}
function statusBadge(s: string): string {
  return {
    pending: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
    running: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20',
    completed: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    failed: 'bg-red-500/10 text-red-400 border-red-500/20',
  }[s] ?? 'bg-zinc-700/50 text-zinc-400 border-zinc-700'
}
</script>

<template>
  <div class="card p-5">
    <div class="flex items-center gap-2 mb-4">
      <h3 class="font-semibold text-zinc-200 text-sm">Browser Research</h3>
      <ProviderChip name="Rtrvr" />
    </div>

    <!-- Run history -->
    <div v-if="runs.length" class="space-y-3 mb-5">
      <div v-for="run in runs" :key="run.id" class="bg-surface-700 rounded-xl p-3">
        <div class="flex items-center gap-2 mb-1.5">
          <span :class="['status-dot flex-shrink-0', statusDot(run.status)]" />
          <span :class="['badge border text-xs', statusBadge(run.status)]">{{ run.status }}</span>
          <span class="text-zinc-600 text-xs ml-auto">{{ run.provider }}</span>
        </div>
        <p class="text-zinc-400 text-xs leading-relaxed">{{ run.task }}</p>
        <p v-if="run.outputKey" class="text-zinc-600 text-xs mt-1.5 font-mono truncate">{{ run.outputKey }}</p>
      </div>
    </div>

    <!-- Ad-hoc research -->
    <form @submit.prevent="runResearch" class="space-y-3">
      <div>
        <label class="block text-xs font-medium text-zinc-500 mb-1.5">Ad-hoc browser research task</label>
        <input
          v-model="customTask"
          type="text"
          placeholder="e.g. Find official documentation for React 18 hooks"
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

    <p v-if="!runs.length && !isRunning" class="text-zinc-700 text-xs mt-3">
      No research runs yet. The workflow will trigger one automatically if Rtrvr is configured.
    </p>
  </div>
</template>
