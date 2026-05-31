<script setup lang="ts">
const emit = defineEmits<{ submit: [url: string, title: string, confirmed: boolean] }>()

const fileUrl = ref('')
const title = ref('')
const confirmed = ref(false)
const error = ref<string | null>(null)

function validate(): boolean {
  error.value = null
  if (!fileUrl.value.trim()) {
    error.value = 'Please enter a file URL.'
    return false
  }
  try {
    const u = new URL(fileUrl.value)
    if (u.protocol !== 'https:') {
      error.value = 'Only HTTPS URLs are allowed.'
      return false
    }
  } catch {
    error.value = 'Enter a valid URL.'
    return false
  }
  if (!confirmed.value) {
    error.value = 'Please confirm you have rights to process this file.'
    return false
  }
  return true
}

function onSubmit() {
  if (validate()) {
    emit('submit', fileUrl.value.trim(), title.value.trim(), confirmed.value)
  }
}
</script>

<template>
  <form @submit.prevent="onSubmit" class="space-y-4">
    <div>
      <label for="fileUrl" class="block text-sm font-medium text-zinc-300 mb-1.5">Direct file URL</label>
      <input
        id="fileUrl"
        v-model="fileUrl"
        type="url"
        placeholder="https://example.com/video.mp4"
        class="input-base"
        autocomplete="off"
      />
      <p class="text-zinc-500 text-xs mt-1.5">Allowed: .mp4, .mov, .webm, .mp3, .wav, .vtt, .srt, .txt, .json</p>
    </div>

    <div>
      <label for="directTitle" class="block text-sm font-medium text-zinc-300 mb-1.5">Title (optional)</label>
      <input
        id="directTitle"
        v-model="title"
        type="text"
        placeholder="My video analysis"
        class="input-base"
      />
    </div>

    <label class="flex items-start gap-3 cursor-pointer">
      <input v-model="confirmed" type="checkbox" class="mt-0.5 rounded border-zinc-600 bg-surface-600" />
      <span class="text-sm text-zinc-400 leading-relaxed">
        I confirm that I own or have explicit permission to process this file. I am not attempting to bypass any copyright protections.
      </span>
    </label>

    <ErrorState v-if="error" :message="error" />

    <button type="submit" class="btn-primary w-full justify-center">
      Start Workflow
    </button>
  </form>
</template>
