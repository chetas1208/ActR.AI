<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const api = useApi()
const workflowStore = useWorkflowStore()

const activeMode = ref<'youtube' | 'upload' | 'direct'>((route.query.mode as string) === 'upload' ? 'upload' : route.query.mode === 'direct' ? 'direct' : 'youtube')

// YouTube mode state
const youtubeUrl = ref('')
const ytSubmitting = ref(false)
const ytError = ref<string | null>(null)

// Upload mode state
const uploadFile = ref<File | null>(null)
const uploadTitle = ref('')
const uploadProgress = ref(0)
const uploadSubmitting = ref(false)
const uploadError = ref<string | null>(null)

// Direct file mode state
const directSubmitting = ref(false)
const directError = ref<string | null>(null)

async function startYouTube() {
  if (!youtubeUrl.value.trim()) return
  ytSubmitting.value = true
  ytError.value = null
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
    ytError.value = e.message
  } finally {
    ytSubmitting.value = false
  }
}

async function handleUploadFile(file: File) {
  uploadFile.value = file
  if (!uploadTitle.value) {
    uploadTitle.value = file.name.replace(/\.[^/.]+$/, '')
  }
}

async function submitUpload() {
  if (!uploadFile.value) return
  uploadSubmitting.value = true
  uploadError.value = null
  uploadProgress.value = 0

  try {
    const { jobId, uploadUrl, objectKey } = await api.presignUpload({
      filename: uploadFile.value.name,
      contentType: uploadFile.value.type || 'application/octet-stream',
      fileSize: uploadFile.value.size,
      title: uploadTitle.value,
    })

    await uploadToTigris(uploadUrl, uploadFile.value, (pct) => {
      uploadProgress.value = pct
    })

    const result = await api.startUploadWorkflow({ jobId, objectKey, title: uploadTitle.value })

    workflowStore.addJob({
      id: jobId,
      sourceType: 'upload',
      sourceRights: 'uploaded_by_user',
      status: result.status as any,
      title: uploadTitle.value || undefined,
      progress: 10,
      requiresUserInput: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })
    router.push(`/workflows/${jobId}`)
  } catch (e: any) {
    uploadError.value = e.message
  } finally {
    uploadSubmitting.value = false
  }
}

async function handleDirectFile(url: string, title: string, confirmed: boolean) {
  directSubmitting.value = true
  directError.value = null
  try {
    const result = await api.startDirectFileWorkflow({ fileUrl: url, title, sourceRightsConfirmed: confirmed })
    workflowStore.addJob({
      id: result.jobId,
      sourceType: 'authorized_direct_file',
      sourceRights: 'authorized_direct_file',
      status: result.status as any,
      title: title || undefined,
      progress: 10,
      requiresUserInput: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })
    router.push(`/workflows/${result.jobId}`)
  } catch (e: any) {
    directError.value = e.message
  } finally {
    directSubmitting.value = false
  }
}

useHead({ title: 'New Workflow – Gorube Flow' })
</script>

<template>
  <AppShell>
    <div class="max-w-2xl mx-auto px-4 sm:px-6 py-12">
      <h1 class="text-2xl font-bold text-zinc-100 mb-2">New Workflow</h1>
      <p class="text-zinc-400 text-sm mb-8">Choose how you want to provide your content.</p>

      <!-- Mode tabs -->
      <div class="flex gap-1 bg-surface-700 rounded-xl p-1 mb-8" role="tablist">
        <button
          v-for="mode in modes"
          :key="mode.value"
          @click="activeMode = mode.value as any"
          :class="[
            'flex-1 flex items-center justify-center gap-2 py-2 px-3 rounded-lg text-sm font-medium transition-all duration-200',
            activeMode === mode.value
              ? 'bg-surface-500 text-zinc-100 shadow-sm'
              : 'text-zinc-500 hover:text-zinc-300',
          ]"
          :aria-selected="activeMode === mode.value"
          role="tab"
        >
          <span>{{ mode.icon }}</span>
          {{ mode.label }}
        </button>
      </div>

      <!-- YouTube mode -->
      <div v-if="activeMode === 'youtube'" class="card p-6 animate-fade-in">
        <h2 class="font-semibold text-zinc-200 text-sm mb-1">YouTube Link</h2>
        <p class="text-zinc-500 text-xs mb-4">Paste any public YouTube URL. Metadata is fetched via the YouTube API. Transcript upload may be required.</p>
        <form @submit.prevent="startYouTube" class="space-y-4">
          <div>
            <label for="ytUrl" class="block text-sm font-medium text-zinc-300 mb-1.5">YouTube URL</label>
            <input
              id="ytUrl"
              v-model="youtubeUrl"
              type="url"
              placeholder="https://www.youtube.com/watch?v=..."
              class="input-base"
              :disabled="ytSubmitting"
              aria-label="YouTube URL"
              required
            />
          </div>
          <ErrorState v-if="ytError" :message="ytError" />
          <button type="submit" :disabled="ytSubmitting || !youtubeUrl.trim()" class="btn-primary w-full justify-center">
            <span v-if="ytSubmitting" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            Start YouTube Workflow
          </button>
        </form>
      </div>

      <!-- Upload mode -->
      <div v-else-if="activeMode === 'upload'" class="card p-6 animate-fade-in">
        <h2 class="font-semibold text-zinc-200 text-sm mb-1">Upload Media</h2>
        <p class="text-zinc-500 text-xs mb-4">Upload a video, audio file, or transcript directly. Files are stored privately in Tigris.</p>

        <div class="space-y-4">
          <UploadDropzone @file="handleUploadFile" :is-loading="uploadSubmitting" />

          <div v-if="uploadFile" class="card-glass p-3 flex items-center gap-2 text-sm">
            <svg class="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
            </svg>
            <span class="text-zinc-300 truncate">{{ uploadFile.name }}</span>
            <span class="text-zinc-500 text-xs ml-auto flex-shrink-0">{{ (uploadFile.size / 1024 / 1024).toFixed(1) }} MB</span>
          </div>

          <div v-if="uploadFile">
            <label for="uploadTitle" class="block text-sm font-medium text-zinc-300 mb-1.5">Title (optional)</label>
            <input id="uploadTitle" v-model="uploadTitle" type="text" placeholder="My video analysis" class="input-base" />
          </div>

          <div v-if="uploadSubmitting" class="space-y-1">
            <div class="w-full bg-surface-600 rounded-full h-1.5">
              <div class="bg-accent-500 h-1.5 rounded-full transition-all" :style="{ width: `${uploadProgress}%` }" />
            </div>
            <p class="text-xs text-zinc-400 text-right">{{ uploadProgress }}% uploaded</p>
          </div>

          <ErrorState v-if="uploadError" :message="uploadError" />

          <button
            v-if="uploadFile"
            @click="submitUpload"
            :disabled="uploadSubmitting"
            class="btn-primary w-full justify-center"
          >
            <span v-if="uploadSubmitting" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            Start Upload Workflow
          </button>
        </div>
      </div>

      <!-- Direct file URL mode -->
      <div v-else-if="activeMode === 'direct'" class="card p-6 animate-fade-in">
        <h2 class="font-semibold text-zinc-200 text-sm mb-1">Authorized Direct File URL</h2>
        <p class="text-zinc-500 text-xs mb-4">Provide a direct HTTPS link to a media file you own or have permission to process. The backend validates and streams it into Tigris.</p>
        <ErrorState v-if="directError" :message="directError" class="mb-4" />
        <DirectFileUrlForm
          :is-loading="directSubmitting"
          @submit="handleDirectFile"
        />
      </div>
    </div>
  </AppShell>
</template>

<script lang="ts">
const modes = [
  { value: 'youtube', icon: '▶', label: 'YouTube Link' },
  { value: 'upload', icon: '⬆', label: 'Upload Media' },
  { value: 'direct', icon: '🔗', label: 'Direct URL' },
]
export default { setup() { return { modes } } }
</script>
