<script setup lang="ts">
const route = useRoute()
const jobId = computed(() => route.params.id as string)
const jobIdRef = computed(() => jobId.value)

const { details, error, status, isTerminal, isFailed, requiresInput, fetch: refetchJob } = useWorkflowPolling(jobIdRef)

const summary = ref<any>(null)
const isLoadingSummary = ref(false)

watch(() => details.value?.artifactUrls?.summary, async (url) => {
  if (!url || summary.value) return
  isLoadingSummary.value = true
  try {
    const res = await fetch(url)
    if (res.ok) summary.value = await res.json()
  } catch {
    // non-fatal
  } finally {
    isLoadingSummary.value = false
  }
})

const workflowStore = useWorkflowStore()
watch(details, (d) => {
  if (d?.job) workflowStore.updateJob(d.job.id, d.job)
})

function onInputDone() {
  refetchJob()
}

function onBrowserRan() {
  refetchJob()
}

const pageTitle = computed(() => {
  const title = details.value?.job?.title
  return title ? `${title} – Gorube Flow` : 'Workflow – Gorube Flow'
})

useHead({ title: pageTitle })
</script>

<template>
  <AppShell>
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading -->
      <div v-if="!details && !error" class="flex items-center justify-center py-20">
        <div class="flex flex-col items-center gap-4">
          <div class="w-8 h-8 border-2 border-accent-400 border-t-transparent rounded-full animate-spin" />
          <p class="text-zinc-400 text-sm">Loading workflow…</p>
        </div>
      </div>

      <ErrorState v-else-if="error && !details" :message="error" title="Failed to load workflow" class="max-w-lg mx-auto mt-16" />

      <template v-else-if="details">
        <!-- Header -->
        <div class="flex items-start justify-between gap-4 mb-8">
          <div>
            <div class="flex items-center gap-2 text-xs text-zinc-500 mb-2">
              <NuxtLink to="/" class="hover:text-zinc-300 transition-colors">Home</NuxtLink>
              <span>/</span>
              <span>Workflows</span>
              <span>/</span>
              <span class="text-zinc-400 font-mono">{{ details.job.id.slice(0, 8) }}…</span>
            </div>
            <h1 class="text-xl font-bold text-zinc-100 leading-tight">
              {{ details.job.title ?? 'Untitled Workflow' }}
            </h1>
            <div class="flex items-center gap-3 mt-2 text-xs text-zinc-500">
              <span class="badge bg-surface-600 text-zinc-400">{{ details.job.sourceType.replace(/_/g, ' ') }}</span>
              <span>{{ new Date(details.job.createdAt).toLocaleDateString() }}</span>
            </div>
          </div>

          <button
            v-if="!isTerminal"
            @click="() => { const api = useApi(); api.continueWorkflow(jobId).then(() => refetchJob()) }"
            class="btn-ghost text-xs flex-shrink-0"
            aria-label="Manually advance workflow step"
          >
            Continue ›
          </button>
        </div>

        <div class="grid lg:grid-cols-5 gap-6">
          <!-- Left column -->
          <div class="lg:col-span-2 space-y-5">
            <VideoPreview
              v-if="details.video"
              :video="details.video"
              :playback-url="details.artifactUrls?.playback"
            />

            <WaitingForInputPanel
              v-if="requiresInput"
              :job-id="jobId"
              :input-type="details.job.requiredInputType"
              @done="onInputDone"
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

            <ErrorState
              v-if="isFailed && details.job.errorMessage"
              :message="details.job.errorMessage"
              title="Workflow failed"
            />
          </div>

          <!-- Right column -->
          <div class="lg:col-span-3 space-y-5">
            <WorkflowTimeline
              :steps="details.steps ?? []"
              :current-status="details.job.status"
              :progress="details.job.progress"
            />

            <!-- Action cards -->
            <div v-if="details.actionCards?.length" class="card p-5">
              <h3 class="font-semibold text-zinc-200 text-sm mb-4">
                Action Cards
                <span class="text-zinc-500 font-normal">({{ details.actionCards.length }})</span>
              </h3>
              <div class="space-y-3">
                <ActionCard
                  v-for="card in details.actionCards"
                  :key="card.id"
                  :card="card"
                  @run="() => refetchJob()"
                />
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
                <span class="text-zinc-500 font-normal">({{ details.claims.length }})</span>
              </h3>
              <div class="space-y-3">
                <ClaimCard v-for="claim in details.claims" :key="claim.id" :claim="claim" />
              </div>
            </div>
            <div v-else-if="!isTerminal && status !== 'ready'" class="card p-5">
              <h3 class="font-semibold text-zinc-200 text-sm mb-3">Claims</h3>
              <LoadingSkeleton :lines="2" />
            </div>

            <!-- Browser Research (Rtrvr) -->
            <BrowserResearchPanel
              :runs="details.browserRuns ?? []"
              :job-id="jobId"
              @ran="onBrowserRan"
            />

            <!-- Execution logs (Daytona) -->
            <ExecutionLogPanel v-if="details.executions?.length" :runs="details.executions" />

            <!-- Provider status hints when keys are missing -->
            <div v-if="status === 'ready'" class="space-y-2">
              <div v-if="!details.executions?.length" class="card p-4 border-zinc-700/40">
                <p class="text-xs text-zinc-500">
                  Provider not configured: add <code class="text-accent-400">DAYTONA_API_KEY</code> to enable safe code execution.
                </p>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppShell>
</template>
