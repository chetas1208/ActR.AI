<script setup lang="ts">
const router = useRouter()
const api = useApi()
const workflowStore = useWorkflowStore()

const youtubeUrl = ref('')
const isSubmitting = ref(false)
const submitError = ref<string | null>(null)

async function startYouTube() {
  if (!youtubeUrl.value.trim() || isSubmitting.value) return
  isSubmitting.value = true
  submitError.value = null

  try {
    const result = await api.startYouTubeWorkflow(youtubeUrl.value.trim())
    workflowStore.addJob({
      id: result.jobId,
      sourceType: 'youtube',
      sourceRights: 'youtube_embed_only',
      status: result.status,
      progress: 10,
      requiresUserInput: result.requiresUpload,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })
    router.push(`/workflows/${result.jobId}`)
  } catch (e: any) {
    submitError.value = e.message
  } finally {
    isSubmitting.value = false
  }
}

useHead({ title: 'Gorube Flow – Turn Videos Into Agent Workflows' })
</script>

<template>
  <AppShell>
    <!-- Hero -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pt-20 pb-12 text-center">
      <!-- Badge -->
      <div class="inline-flex items-center gap-2 badge bg-accent-500/10 text-accent-300 border border-accent-500/20 mb-6 text-xs">
        <span class="status-dot bg-accent-500 animate-pulse" />
        Built on GoTube · Powered by Tigris &amp; Rtrvr
      </div>

      <!-- Headline -->
      <h1 class="text-4xl sm:text-5xl font-bold text-zinc-100 tracking-tight mb-4 leading-tight">
        Turn videos into<br />
        <span class="text-transparent bg-clip-text bg-gradient-to-r from-accent-400 to-violet-400">workflows that act.</span>
      </h1>
      <p class="text-zinc-400 text-lg max-w-2xl mx-auto mb-10">
        Paste a YouTube link, upload media, or connect a direct file URL. Gorube Flow extracts actions, verifies claims, and safely executes code steps.
      </p>

      <!-- Main input -->
      <form @submit.prevent="startYouTube" class="flex flex-col sm:flex-row gap-3 max-w-xl mx-auto mb-4">
        <input
          v-model="youtubeUrl"
          type="url"
          placeholder="https://www.youtube.com/watch?v=..."
          class="input-base flex-1 text-sm"
          :disabled="isSubmitting"
          aria-label="YouTube URL"
        />
        <button type="submit" :disabled="isSubmitting || !youtubeUrl.trim()" class="btn-primary whitespace-nowrap">
          <span v-if="isSubmitting" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
          <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
          </svg>
          Start Workflow
        </button>
      </form>

      <ErrorState v-if="submitError" :message="submitError" class="max-w-xl mx-auto mb-4" />

      <!-- Other modes -->
      <div class="flex flex-col sm:flex-row items-center justify-center gap-3">
        <NuxtLink to="/new?mode=upload" class="btn-secondary text-sm">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
          </svg>
          Upload Media
        </NuxtLink>
        <NuxtLink to="/new?mode=direct" class="btn-secondary text-sm">
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
          </svg>
          Direct File URL
        </NuxtLink>
      </div>
    </section>

    <!-- How it works -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pb-16">
      <div class="card-glass p-1 rounded-2xl">
        <div class="grid grid-cols-5 gap-0 text-center text-xs">
          <div v-for="(step, i) in howItWorks" :key="step.label" class="flex flex-col items-center py-4 px-3 relative">
            <div class="w-8 h-8 rounded-lg bg-surface-600 flex items-center justify-center mb-2 text-base">{{ step.icon }}</div>
            <span class="font-medium text-zinc-300">{{ step.label }}</span>
            <div v-if="i < 4" class="absolute right-0 top-1/2 -translate-y-1/2 text-zinc-600 text-xs">›</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Feature highlights -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pb-20">
      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="feat in features" :key="feat.title" class="card p-4">
          <div class="text-2xl mb-2">{{ feat.icon }}</div>
          <h3 class="font-semibold text-zinc-200 text-sm mb-1">{{ feat.title }}</h3>
          <p class="text-zinc-500 text-xs leading-relaxed">{{ feat.description }}</p>
          <SponsorBadge :name="feat.sponsor" :color="feat.color" class="mt-3" />
        </div>
      </div>
    </section>

    <!-- Recent workflows -->
    <section v-if="workflowStore.recentJobs.length" class="max-w-4xl mx-auto px-4 sm:px-6 pb-20">
      <h2 class="font-semibold text-zinc-300 text-sm mb-4">Recent Workflows</h2>
      <div class="space-y-2">
        <NuxtLink
          v-for="job in workflowStore.recentJobs"
          :key="job.id"
          :to="`/workflows/${job.id}`"
          class="card p-3 flex items-center gap-3 hover:border-zinc-700/80 transition-colors"
        >
          <div :class="['status-dot flex-shrink-0', statusDot(job.status)]" />
          <span class="text-zinc-200 text-sm flex-1 truncate">{{ job.title ?? job.id }}</span>
          <span class="badge bg-surface-600 text-zinc-400 text-xs">{{ job.sourceType }}</span>
          <span class="text-zinc-500 text-xs">{{ job.status }}</span>
        </NuxtLink>
      </div>
    </section>
  </AppShell>
</template>

<script lang="ts">
const howItWorks = [
  { icon: '📥', label: 'Input' },
  { icon: '🔍', label: 'Extract' },
  { icon: '✓', label: 'Verify' },
  { icon: '⚡', label: 'Act' },
  { icon: '📤', label: 'Export' },
]

const features = [
  { icon: '🃏', title: 'Action Cards', description: 'Structured actions extracted from video content. Code, research, checklists, study plans.', sponsor: 'Daytona', color: 'purple' },
  { icon: '🔎', title: 'Source Research', description: 'Claims and sources researched via Rtrvr browser automation.', sponsor: 'Rtrvr', color: 'pink' },
  { icon: '🛡️', title: 'Safe Code Execution', description: 'Generated code runs in isolated Daytona sandboxes with captured logs.', sponsor: 'Daytona', color: 'purple' },
  { icon: '📦', title: 'Artifact Storage', description: 'Every artifact stored in Tigris with signed URLs.', sponsor: 'Tigris', color: 'blue' },
]

function statusDot(status: string): string {
  const map: Record<string, string> = {
    ready: 'bg-emerald-500',
    failed: 'bg-red-500',
    waiting_for_user_input: 'bg-violet-500',
  }
  return map[status] ?? 'bg-blue-500 animate-pulse'
}

export default { setup() { return { howItWorks, features, statusDot } } }
</script>
