<script setup lang="ts">
const router = useRouter()
const api    = useApi()
const workflowStore = useWorkflowStore()

const activeTab = ref<'youtube' | 'upload' | 'direct'>('youtube')
const youtubeUrl = ref('')
const isSubmitting = ref(false)
const submitError  = ref<string | null>(null)

async function startYouTube() {
  if (!youtubeUrl.value.trim() || isSubmitting.value) return
  isSubmitting.value = true
  submitError.value  = null
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

useHead({ title: 'ActR.AI – Video to Agent Workflow' })
</script>

<template>
  <AppShell>
    <!-- Hero -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pt-24 pb-10 text-center">
      <!-- Eyebrow -->
      <div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-cyan-500/10 border border-cyan-500/20 mb-8 text-xs text-cyan-400 font-medium">
        <span class="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse" />
        Powered by NVIDIA NIM · Tigris · Daytona · Rtrvr
      </div>

      <!-- Headline -->
      <h1 class="text-5xl sm:text-6xl font-bold tracking-tight mb-5 leading-tight">
        <span class="text-zinc-100">ActR.AI turns</span><br />
        <span class="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-electric-400 to-violet-400">
          video into action.
        </span>
      </h1>

      <p class="text-zinc-400 text-lg max-w-2xl mx-auto mb-10 leading-relaxed">
        Paste a YouTube link, upload media, or connect a direct file URL.
        ActR.AI extracts actions, researches sources, and safely executes code steps.
      </p>

      <!-- Source input module -->
      <div class="card-glass max-w-2xl mx-auto mb-4 overflow-hidden">
        <!-- Tabs -->
        <div class="flex border-b border-zinc-800/60">
          <button
            v-for="tab in tabs"
            :key="tab.value"
            @click="activeTab = tab.value as any"
            :class="[
              'flex-1 flex items-center justify-center gap-1.5 py-3 px-4 text-sm font-medium transition-all duration-200',
              activeTab === tab.value
                ? 'text-cyan-400 border-b-2 border-cyan-400 bg-cyan-500/5'
                : 'text-zinc-500 hover:text-zinc-300 border-b-2 border-transparent',
            ]"
          >
            <span>{{ tab.icon }}</span>
            <span class="hidden sm:inline">{{ tab.label }}</span>
          </button>
        </div>

        <!-- YouTube tab -->
        <div v-if="activeTab === 'youtube'" class="p-5">
          <form @submit.prevent="startYouTube" class="flex gap-2">
            <input
              v-model="youtubeUrl"
              type="url"
              placeholder="https://www.youtube.com/watch?v=..."
              class="input-base flex-1 text-sm"
              :disabled="isSubmitting"
              aria-label="YouTube URL"
            />
            <button
              type="submit"
              :disabled="isSubmitting || !youtubeUrl.trim()"
              class="btn-primary whitespace-nowrap"
            >
              <span v-if="isSubmitting" class="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
              <span v-else>Start Workflow</span>
            </button>
          </form>
        </div>

        <!-- Upload tab -->
        <div v-else-if="activeTab === 'upload'" class="p-5 text-center">
          <p class="text-zinc-400 text-sm mb-3">Upload a video, audio, or transcript file</p>
          <NuxtLink to="/new?mode=upload" class="btn-primary">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
            </svg>
            Choose File
          </NuxtLink>
        </div>

        <!-- Direct URL tab -->
        <div v-else-if="activeTab === 'direct'" class="p-5 text-center">
          <p class="text-zinc-400 text-sm mb-3">Provide a direct HTTPS link to a file you own</p>
          <NuxtLink to="/new?mode=direct" class="btn-primary">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
            </svg>
            Enter URL
          </NuxtLink>
        </div>
      </div>

      <ErrorState v-if="submitError" :message="submitError" class="max-w-2xl mx-auto mb-4" />
    </section>

    <!-- How it works -->
    <section class="max-w-3xl mx-auto px-4 sm:px-6 pb-16">
      <div class="flex items-center justify-center gap-2 sm:gap-4 text-xs sm:text-sm">
        <div
          v-for="(step, i) in howItWorks"
          :key="step.label"
          class="flex items-center gap-2 sm:gap-4"
        >
          <div class="flex flex-col items-center gap-1.5">
            <div class="w-9 h-9 rounded-xl bg-surface-700 border border-zinc-700/50 flex items-center justify-center text-base">
              {{ step.icon }}
            </div>
            <span class="text-zinc-500 text-xs">{{ step.label }}</span>
          </div>
          <svg v-if="i < howItWorks.length - 1" class="w-4 h-4 text-zinc-700 flex-shrink-0 hidden sm:block" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </div>
      </div>
    </section>

    <!-- Feature grid -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pb-20">
      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="feat in features" :key="feat.title" class="card p-4 hover:border-zinc-700/80 transition-colors">
          <div class="text-2xl mb-2">{{ feat.icon }}</div>
          <h3 class="font-semibold text-zinc-200 text-sm mb-1">{{ feat.title }}</h3>
          <p class="text-zinc-500 text-xs leading-relaxed">{{ feat.description }}</p>
          <ProviderChip :name="feat.provider" class="mt-3" />
        </div>
      </div>
    </section>

    <!-- Recent workflows -->
    <section v-if="workflowStore.recentJobs.length" class="max-w-4xl mx-auto px-4 sm:px-6 pb-20">
      <h2 class="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">Recent Workflows</h2>
      <div class="space-y-2">
        <NuxtLink
          v-for="job in workflowStore.recentJobs"
          :key="job.id"
          :to="`/workflows/${job.id}`"
          class="card p-3 flex items-center gap-3 hover:border-zinc-700/80 transition-colors"
        >
          <div :class="['status-dot flex-shrink-0', statusDot(job.status)]" />
          <span class="text-zinc-200 text-sm flex-1 truncate">{{ job.title || job.id }}</span>
          <span class="badge bg-surface-600 text-zinc-400 text-xs">{{ job.sourceType }}</span>
          <span class="text-zinc-500 text-xs hidden sm:block">{{ job.status }}</span>
        </NuxtLink>
      </div>
    </section>
  </AppShell>
</template>

<script lang="ts">
const tabs = [
  { value: 'youtube', icon: '▶', label: 'YouTube Link' },
  { value: 'upload',  icon: '⬆', label: 'Upload File' },
  { value: 'direct',  icon: '🔗', label: 'Direct URL' },
]
const howItWorks = [
  { icon: '📥', label: 'Input' },
  { icon: '🔍', label: 'Extract' },
  { icon: '🌐', label: 'Research' },
  { icon: '⚡', label: 'Act' },
  { icon: '📤', label: 'Export' },
]
const features = [
  { icon: '🃏', title: 'Action Cards',     description: 'Structured actions extracted from any video via NVIDIA NIM reasoning.', provider: 'NVIDIA NIM' },
  { icon: '🌐', title: 'Source Research',  description: 'Claims and sources verified through Rtrvr browser automation.',           provider: 'Rtrvr' },
  { icon: '🛡️', title: 'Safe Execution',   description: 'Generated code runs in isolated Daytona sandboxes with full audit logs.',  provider: 'Daytona' },
  { icon: '📦', title: 'Artifact Storage', description: 'Every artifact stored privately in Tigris with signed access URLs.',        provider: 'Tigris' },
]
function statusDot(status: string): string {
  const m: Record<string, string> = {
    ready: 'bg-emerald-500',
    failed: 'bg-red-500',
    waiting_for_user_input: 'bg-violet-500',
  }
  return m[status] ?? 'bg-cyan-500 animate-pulse'
}
export default { setup() { return { tabs, howItWorks, features, statusDot } } }
</script>
