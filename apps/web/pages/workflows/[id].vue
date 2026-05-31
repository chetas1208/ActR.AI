<script setup lang="ts">
import { STATUS_LABELS, STATUS_COLOR } from '~/types'

const route = useRoute()
const jobId = computed(() => route.params.id as string)
const jobIdRef = computed(() => jobId.value)

const { details, error, status, isTerminal, isFailed, requiresInput, progress, fetch: refetchJob } = useWorkflowPolling(jobIdRef)

const summary = ref<any>(null)
const isLoadingSummary = ref(false)

watch(() => details.value?.artifactUrls?.summary, async (url) => {
  if (!url || summary.value) return
  isLoadingSummary.value = true
  try {
    const res = await fetch(url)
    if (res.ok) summary.value = await res.json()
  } catch { /* non-fatal */ }
  finally { isLoadingSummary.value = false }
})

const workflowStore = useWorkflowStore()
watch(details, d => { if (d?.job) workflowStore.updateJob(d.job.id, d.job) })

const api = useApi()
async function manualContinue() {
  try { await api.continueWorkflow(jobId.value); await refetchJob() } catch { /* ignore */ }
}

const pageTitle = computed(() => {
  const t = details.value?.job?.title
  return t ? `${t} – ActR.AI` : 'Workflow – ActR.AI'
})
useHead({ title: pageTitle })
</script>

<template>
  <AppShell>
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">

      <!-- Loading -->
      <div v-if="!details && !error" class="flex items-center justify-center py-24">
        <div class="flex flex-col items-center gap-4">
          <div class="w-10 h-10 border-2 border-cyan-400 border-t-transparent rounded-full animate-spin" />
          <p class="text-zinc-400 text-sm">Loading workflow…</p>
        </div>
      </div>

      <ErrorState v-else-if="error && !details" :message="error" title="Failed to load workflow" class="max-w-lg mx-auto mt-16" />

      <template v-else-if="details">
        <!-- Header bar -->
        <div class="flex items-start justify-between gap-4 mb-8">
          <div class="min-w-0">
            <div class="flex items-center gap-2 text-xs text-zinc-600 mb-2">
              <NuxtLink to="/" class="hover:text-zinc-400 transition-colors">ActR.AI</NuxtLink>
              <span>/</span>
              <span>Workflows</span>
              <span>/</span>
              <span class="font-mono text-zinc-500">{{ details.job.id.slice(0, 8) }}…</span>
            </div>
            <h1 class="text-xl font-bold text-zinc-100 truncate">{{ details.job.title ?? 'Untitled Workflow' }}</h1>
            <div class="flex items-center gap-2 mt-2 flex-wrap">
              <span class="provider-chip">{{ details.job.sourceType.replace(/_/g, ' ') }}</span>
              <span :class="['text-xs font-medium', STATUS_COLOR[details.job.status]]">
                {{ STATUS_LABELS[details.job.status] ?? details.job.status }}
              </span>
            </div>
          </div>

          <button v-if="!isTerminal" @click="manualContinue" class="btn-ghost text-xs flex-shrink-0 whitespace-nowrap">
            Continue ›
          </button>
        </div>

        <!-- Two-column layout -->
        <div class="grid lg:grid-cols-5 gap-6">
          <!-- Left: source + waiting panel -->
          <div class="lg:col-span-2 space-y-5">
            <VideoPreview v-if="details.video" :video="details.video" :playback-url="details.artifactUrls?.playback" />

            <WaitingForInputPanel
              v-if="requiresInput"
              :job-id="jobId"
              :input-type="details.job.requiredInputType"
              @done="refetchJob"
            />

            <SummaryCard
              v-if="summary || isLoadingSummary"
              :title="summary?.title ?? ''"
              :summary="summary?.summary ?? ''"
              :audience="summary?.audience"
              :difficulty="summary?.difficulty"
              :key-topics="summary?.keyTopics"
              :is-loading="isLoadingSummary"
            />

            <ErrorState v-if="isFailed && details.job.errorMessage" :message="details.job.errorMessage" title="Workflow failed" />
          </div>

          <!-- Right: timeline + results -->
          <div class="lg:col-span-3 space-y-5">
            <!-- Timeline -->
            <WorkflowTimeline :steps="details.steps ?? []" :current-status="details.job.status" :progress="progress" />

            <!-- Action cards -->
            <div v-if="details.actionCards?.length" class="card p-5">
              <h3 class="font-semibold text-zinc-200 text-sm mb-4">
                Action Cards
                <span class="text-zinc-500 font-normal ml-1">({{ details.actionCards.length }})</span>
              </h3>
              <div class="space-y-3">
                <ActionCard v-for="card in details.actionCards" :key="card.id" :card="card" @run="refetchJob" />
              </div>
            </div>
            <div v-else-if="!isTerminal" class="card p-5">
              <h3 class="font-semibold text-zinc-200 text-sm mb-3">Action Cards</h3>
              <LoadingSkeleton :lines="3" />
            </div>

            <!-- Claims -->
            <div v-if="details.claims?.length" class="card p-5">
              <h3 class="font-semibold text-zinc-200 text-sm mb-4">
                Claims
                <span class="text-zinc-500 font-normal ml-1">({{ details.claims.length }})</span>
              </h3>
              <div class="space-y-3">
                <ClaimCard v-for="claim in details.claims" :key="claim.id" :claim="claim" />
              </div>
            </div>
            <div v-else-if="!isTerminal" class="card p-5">
              <h3 class="font-semibold text-zinc-200 text-sm mb-3">Claims</h3>
              <LoadingSkeleton :lines="2" />
            </div>

            <!-- Browser Research (Rtrvr) -->
            <BrowserResearchPanel :runs="details.browserRuns ?? []" :job-id="jobId" @ran="refetchJob" />

            <!-- Execution Logs (Daytona) -->
            <ExecutionLogPanel v-if="details.executions?.length" :runs="details.executions" />

            <!-- Provider hints when keys not configured -->
            <div v-if="status === 'ready'" class="space-y-2">
              <div v-if="!details.executions?.length" class="card p-4 border-zinc-700/30">
                <p class="text-xs text-zinc-600">
                  Add <code class="text-cyan-400">DAYTONA_API_KEY</code> to enable safe code execution.
                </p>
              </div>
              <div v-if="!details.browserRuns?.length" class="card p-4 border-zinc-700/30">
                <p class="text-xs text-zinc-600">
                  Add <code class="text-violet-400">RTRVR_API_KEY</code> to enable browser research.
                </p>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppShell>
</template>
