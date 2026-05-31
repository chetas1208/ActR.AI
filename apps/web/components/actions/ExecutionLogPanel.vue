<script setup lang="ts">
import type { ExecutionRun } from '~/types'

defineProps<{ runs: ExecutionRun[] }>()
</script>

<template>
  <div class="card p-5">
    <h3 class="font-semibold text-zinc-200 text-sm mb-4">Execution Logs</h3>
    <div v-if="!runs.length" class="text-zinc-500 text-sm">No executions yet.</div>
    <div v-else class="space-y-3">
      <div
        v-for="run in runs"
        :key="run.id"
        class="bg-surface-700 rounded-xl p-3 font-mono text-xs"
      >
        <div class="flex items-center gap-2 mb-2">
          <span class="text-zinc-400">{{ run.provider }}</span>
          <span :class="['badge text-xs', run.status === 'completed' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400']">
            {{ run.status }}
          </span>
          <span v-if="run.exitCode != null" class="text-zinc-500">exit {{ run.exitCode }}</span>
        </div>
        <p v-if="run.logsKey" class="text-zinc-500">Logs stored at: {{ run.logsKey }}</p>
        <p v-else class="text-zinc-600">No log output captured.</p>
      </div>
    </div>
  </div>
</template>
