import type { WorkflowDetails, WorkflowStatus } from '~/types'
import { TERMINAL_STATUSES } from '~/types'

export function useWorkflowPolling(jobId: Ref<string | null>) {
  const config  = useRuntimeConfig()
  const api     = useApi()
  const details = ref<WorkflowDetails | null>(null)
  const error   = ref<string | null>(null)
  const isPolling = ref(false)
  let pollTimer: ReturnType<typeof setTimeout> | null = null

  const status        = computed<WorkflowStatus | null>(() => details.value?.job?.status ?? null)
  const isTerminal    = computed(() => status.value ? TERMINAL_STATUSES.includes(status.value) : false)
  const isReady       = computed(() => status.value === 'ready')
  const isFailed      = computed(() => status.value === 'failed')
  const requiresInput = computed(() => details.value?.job?.requiresUserInput ?? false)
  const progress      = computed(() => details.value?.job?.progress ?? 0)

  async function fetch() {
    if (!jobId.value) return
    try {
      details.value = await api.getWorkflow(jobId.value)
      error.value   = null
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function advance() {
    if (!jobId.value || isTerminal.value || requiresInput.value) return
    try {
      await api.continueWorkflow(jobId.value)
    } catch { /* non-fatal */ }
  }

  function schedule() {
    if (pollTimer) clearTimeout(pollTimer)
    const interval = Number(config.public.workflowPollIntervalMs) || 2000
    pollTimer = setTimeout(async () => {
      if (!isPolling.value || !jobId.value) return
      await fetch()
      if (!isTerminal.value) {
        if (config.public.workflowAutoAdvance) await advance()
        schedule()
      }
    }, interval)
  }

  function startPolling() {
    if (isPolling.value) return
    isPolling.value = true
    fetch().then(() => { if (!isTerminal.value) schedule() })
  }

  function stopPolling() {
    isPolling.value = false
    if (pollTimer) { clearTimeout(pollTimer); pollTimer = null }
  }

  watch(jobId, id => { if (id) startPolling(); else stopPolling() }, { immediate: true })
  onUnmounted(() => stopPolling())

  return { details, error, status, isTerminal, isReady, isFailed, requiresInput, progress, fetch, startPolling, stopPolling, advance }
}
