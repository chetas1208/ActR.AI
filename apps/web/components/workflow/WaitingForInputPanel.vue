<script setup lang="ts">
const props = defineProps<{
  jobId: string
  inputType?: string
}>()

const emit = defineEmits<{ done: [] }>()

const api = useApi()
const uploadProgress = ref(0)
const isUploading = ref(false)
const uploadError = ref<string | null>(null)
const uploadDone = ref(false)

async function handleFile(file: File) {
  isUploading.value = true
  uploadError.value = null
  uploadProgress.value = 0

  try {
    const { uploadUrl, objectKey } = await api.presignTranscript({
      jobId: props.jobId,
      filename: file.name,
      contentType: file.type || 'text/plain',
    })

    await uploadToTigris(uploadUrl, file, (pct) => {
      uploadProgress.value = pct
    })

    // Start workflow with the uploaded transcript
    await api.startUploadWorkflow({
      jobId: props.jobId,
      objectKey,
    })

    uploadDone.value = true
    emit('done')
  } catch (e: any) {
    uploadError.value = e.message
  } finally {
    isUploading.value = false
  }
}
</script>

<template>
  <div class="card p-6 border-violet-500/20 bg-violet-500/5">
    <div class="flex items-start gap-3 mb-5">
      <div class="w-8 h-8 rounded-lg bg-violet-500/20 flex items-center justify-center flex-shrink-0">
        <svg class="w-4 h-4 text-violet-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 8.25H7.5a2.25 2.25 0 0 0-2.25 2.25v9a2.25 2.25 0 0 0 2.25 2.25h9a2.25 2.25 0 0 0 2.25-2.25v-9a2.25 2.25 0 0 0-2.25-2.25H15M9 12l3 3m0 0 3-3m-3 3V2.25" />
        </svg>
      </div>
      <div>
        <p class="font-semibold text-violet-300 text-sm">Input Required</p>
        <p class="text-zinc-400 text-sm mt-1">
          <template v-if="inputType === 'transcript_audio_or_video'">
            A transcript could not be automatically obtained. Upload a transcript (.vtt, .srt, .txt), audio, or video file to continue processing.
          </template>
          <template v-else-if="inputType === 'transcript'">
            Please upload a transcript file (.vtt, .srt, .txt) to continue.
          </template>
          <template v-else>
            Additional input is required to continue this workflow.
          </template>
        </p>
      </div>
    </div>

    <template v-if="!uploadDone">
      <UploadDropzone
        :is-loading="isUploading"
        label="Upload transcript, audio, or video"
        @file="handleFile"
      />

      <div v-if="isUploading" class="mt-3">
        <div class="w-full bg-surface-600 rounded-full h-1.5">
          <div
            class="bg-accent-500 h-1.5 rounded-full transition-all"
            :style="{ width: `${uploadProgress}%` }"
          />
        </div>
        <p class="text-zinc-400 text-xs mt-1 text-right">{{ uploadProgress }}%</p>
      </div>

      <ErrorState v-if="uploadError" :message="uploadError" class="mt-3" />
    </template>

    <div v-else class="flex items-center gap-2 text-emerald-400 text-sm">
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
      </svg>
      File uploaded. Workflow continuing…
    </div>
  </div>
</template>
