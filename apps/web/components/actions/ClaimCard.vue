<script setup lang="ts">
import type { Claim } from '~/types'

defineProps<{ claim: Claim }>()

function statusConfig(status: string) {
  const map: Record<string, { label: string; cls: string; dot: string }> = {
    pending:    { label: 'Pending',    cls: 'bg-amber-500/10 text-amber-400 border-amber-500/20',       dot: 'bg-amber-500' },
    researched: { label: 'Researched', cls: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20', dot: 'bg-emerald-500' },
    uncertain:  { label: 'Uncertain',  cls: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',          dot: 'bg-zinc-500' },
    failed:     { label: 'Failed',     cls: 'bg-red-500/10 text-red-400 border-red-500/20',             dot: 'bg-red-500' },
  }
  return map[status] ?? map.pending
}
</script>

<template>
  <div class="card p-4">
    <div class="flex items-start gap-3">
      <div :class="['status-dot flex-shrink-0 mt-1.5', statusConfig(claim.verificationStatus).dot]" />
      <div class="flex-1 min-w-0">
        <p class="text-zinc-300 text-sm leading-relaxed mb-2">{{ claim.claimText }}</p>
        <div class="flex items-center gap-2 flex-wrap">
          <span :class="['badge border text-xs', statusConfig(claim.verificationStatus).cls]">
            {{ statusConfig(claim.verificationStatus).label }}
          </span>
          <span v-if="claim.confidence != null" class="text-xs text-zinc-500">
            {{ Math.round(claim.confidence * 100) }}% confidence
          </span>
          <span v-if="claim.timestampSeconds" class="text-xs text-zinc-500">
            @ {{ Math.floor(claim.timestampSeconds / 60) }}:{{ String(claim.timestampSeconds % 60).padStart(2, '0') }}
          </span>
          <a
            v-if="claim.evidenceKey"
            :href="`#evidence-${claim.id}`"
            class="text-xs text-accent-400 hover:text-accent-300"
          >View evidence ›</a>
        </div>
      </div>
    </div>
  </div>
</template>
