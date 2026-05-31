<script setup lang="ts">
const route  = useRoute()
const router = useRouter()
const api    = useApi()
const workflowStore = useWorkflowStore()

const activeMode = ref<'youtube' | 'upload' | 'direct'>(
  route.query.mode === 'upload' ? 'upload' : route.query.mode === 'direct' ? 'direct' : 'youtube',
)

// YouTube
const youtubeUrl   = ref('')
const ytSubmitting = ref(false)
const ytError      = ref<string | null>(null)

async function startYouTube() {
  if (!youtubeUrl.value.trim()) return
  ytSubmitting.value = true; ytError.value = null
  try {
    const result = await api.startYouTubeWorkflow(youtubeUrl.value.trim())
    workflowStore.addJob({ id: result.jobId, sourceType: 'youtube', sourceRights: 'youtube_embed_only', status: result.status, progress: 10, requiresUserInput: result.requiresUpload, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() })
    router.push(`/workflows/${result.jobId}`)
  } catch (e: any) { ytError.value = e.message }
  finally { ytSubmitting.value = false }
}

// Upload
const uploadFile      = ref<File | null>(null)
const uploadTitle     = ref('')
const uploadProgress  = ref(0)
const uploadSubmitting = ref(false)
const uploadError     = ref<string | null>(null)

function handleUploadFile(file: File) {
  uploadFile.value = file
  if (!uploadTitle.value) uploadTitle.value = file.name.replace(/\.[^/.]+$/, '')
}

async function submitUpload() {
  if (!uploadFile.value) return
  uploadSubmitting.value = true; uploadError.value = null; uploadProgress.value = 0
  try {
    const { jobId, uploadUrl, objectKey } = await api.presignUpload({
      filename: uploadFile.value.name,
      contentType: uploadFile.value.type || 'application/octet-stream',
      fileSize: uploadFile.value.size,
      title: uploadTitle.value,
    })
    await uploadToTigris(uploadUrl, uploadFile.value, p => { uploadProgress.value = p })
    const result = await api.startUploadWorkflow({ jobId, objectKey, title: uploadTitle.value })
    workflowStore.addJob({ id: jobId, sourceType: 'upload', sourceRights: 'uploaded_by_user', status: result.status as any, title: uploadTitle.value || undefined, progress: 10, requiresUserInput: false, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() })
    router.push(`/workflows/${jobId}`)
  } catch (e: any) { uploadError.value = e.message }
  finally { uploadSubmitting.value = false }
}

// Direct URL
const directSubmitting = ref(false)
const directError      = ref<string | null>(null)

async function handleDirectFile(url: string, title: string, confirmed: boolean) {
  directSubmitting.value = true; directError.value = null
  try {
    const result = await api.startDirectFileWorkflow({ fileUrl: url, title, sourceRightsConfirmed: confirmed })
    workflowStore.addJob({ id: result.jobId, sourceType: 'authorized_direct_file', sourceRights: 'authorized_direct_file', status: result.status as any, title: title || undefined, progress: 10, requiresUserInput: false, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() })
    router.push(`/workflows/${result.jobId}`)
  } catch (e: any) { directError.value = e.message }
  finally { directSubmitting.value = false }
}

useHead({ title: 'New Workflow – ActR.AI' })
</script>

<template>
  <AppShell>
    <div class="max-w-xl mx-auto px-4 sm:px-6 py-12">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-2xl font-bold text-zinc-100 mb-1">New Workflow</h1>
        <p class="text-zinc-400 text-sm">Choose your content source to begin.</p>
      </div>

      <!-- Mode tabs -->
      <div class="flex gap-1 bg-surface-800 rounded-xl p-1 mb-6 border border-zinc-800/60">
        <button
          v-for="mode in modes"
          :key="mode.value"
          @click="activeMode = mode.value as any"
          :class="[
            'flex-1 flex items-center justify-center gap-2 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200',
            activeMode === mode.value
              ? 'bg-surface-600 text-zinc-100 shadow-sm border border-zinc-700/60'
              : 'text-zinc-500 hover:text-zinc-300',
          ]"
          role="tab"
          :aria-selected="activeMode === mode.value"
        >
          <span>{{ mode.icon }}</span>
          {{ mode.label }}
        </button>
      </div>

      <!-- YouTube -->
      <div v-if="activeMode === 'youtube'" class="card p-6 animate-fade-in">
        <h2 class="font-semibold text-zinc-200 text-sm mb-1">YouTube Link</h2>
        <p class="text-zinc-500 text-xs mb-4">Paste a public YouTube URL. Metadata is fetched via the YouTube API. Transcript upload may be required.</p>
        <form @submit.prevent="startYouTube" class="space-y-4">
          <div>
            <label for="ytUrl" class="block text-xs font-medium text-zinc-400 mb-1.5">YouTube URL</label>
            <input id="ytUrl" v-model="youtubeUrl" type="url" placeholder="https://www.youtube.com/watch?v=..." class="input-base" :disabled="ytSubmitting" required />
          </div>
          <ErrorState v-if="ytError" :message="ytError" />
          <button type="submit" :disabled="ytSubmitting || !youtubeUrl.trim()" class="btn-primary w-full justify-center">
            <span v-if="ytSubmitting" class="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
            <span v-else>Start Workflow</span>
          </button>
        </form>
      </div>

      <!-- Upload -->
      <div v-else-if="activeMode === 'upload'" class="card p-6 animate-fade-in">
        <h2 class="font-semibold text-zinc-200 text-sm mb-1">Upload Media</h2>
        <p class="text-zinc-500 text-xs mb-4">Upload a video, audio, or transcript file. It is stored privately in Tigris.</p>
        <div class="space-y-4">
          <UploadDropzone @file="handleUploadFile" :is-loading="uploadSubmitting" />
          <div v-if="uploadFile" class="flex items-center gap-2 text-sm bg-surface-700 rounded-xl px-3 py-2">
            <svg class="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
            </svg>
            <span class="text-zinc-300 truncate text-xs">{{ uploadFile.name }}</span>
            <span class="text-zinc-500 text-xs ml-auto flex-shrink-0">{{ (uploadFile.size / 1024 / 1024).toFixed(1) }} MB</span>
          </div>
          <div v-if="uploadFile">
            <label for="uploadTitle" class="block text-xs font-medium text-zinc-400 mb-1.5">Title (optional)</label>
            <input id="uploadTitle" v-model="uploadTitle" type="text" placeholder="My video analysis" class="input-base text-sm" />
          </div>
          <div v-if="uploadSubmitting" class="space-y-1">
            <div class="w-full bg-surface-600 rounded-full h-1.5">
              <div class="bg-gradient-to-r from-cyan-500 to-electric-500 h-1.5 rounded-full transition-all" :style="{ width: `${uploadProgress}%` }" />
            </div>
            <p class="text-xs text-zinc-400 text-right">{{ uploadProgress }}%</p>
          </div>
          <ErrorState v-if="uploadError" :message="uploadError" />
          <button v-if="uploadFile" @click="submitUpload" :disabled="uploadSubmitting" class="btn-primary w-full justify-center">
            <span v-if="uploadSubmitting" class="w-4 h-4 border-2 border-white/40 border-t-white rounded-full animate-spin" />
            <span v-else>Start Workflow</span>
          </button>
        </div>
      </div>

      <!-- Direct URL -->
      <div v-else-if="activeMode === 'direct'" class="card p-6 animate-fade-in">
        <h2 class="font-semibold text-zinc-200 text-sm mb-1">Authorized Direct File URL</h2>
        <p class="text-zinc-500 text-xs mb-4">Provide a direct HTTPS link to a file you own or have permission to process.</p>
        <ErrorState v-if="directError" :message="directError" class="mb-4" />
        <DirectFileUrlForm :is-loading="directSubmitting" @submit="handleDirectFile" />
      </div>
    </div>
  </AppShell>
</template>

<script lang="ts">
const modes = [
  { value: 'youtube', icon: '▶', label: 'YouTube' },
  { value: 'upload',  icon: '⬆', label: 'Upload' },
  { value: 'direct',  icon: '🔗', label: 'Direct URL' },
]
export default { setup() { return { modes } } }
</script>
