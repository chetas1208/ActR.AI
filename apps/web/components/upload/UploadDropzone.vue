<script setup lang="ts">
const props = defineProps<{
  accept?: string
  label?: string
  isLoading?: boolean
}>()

const emit = defineEmits<{ file: [file: File] }>()

const isDragging = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

function onDrop(e: DragEvent) {
  isDragging.value = false
  const file = e.dataTransfer?.files[0]
  if (file) emit('file', file)
}

function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) emit('file', file)
}

function openPicker() {
  fileInput.value?.click()
}
</script>

<template>
  <div
    :class="[
      'border-2 border-dashed rounded-2xl p-8 text-center cursor-pointer transition-all duration-200',
      isDragging ? 'border-accent-400 bg-accent-500/5' : 'border-zinc-700/60 hover:border-zinc-600',
    ]"
    @click="openPicker"
    @dragover.prevent="isDragging = true"
    @dragleave="isDragging = false"
    @drop.prevent="onDrop"
    role="button"
    :aria-label="label ?? 'Upload file'"
  >
    <input
      ref="fileInput"
      type="file"
      :accept="accept ?? 'video/*,audio/*,.vtt,.srt,.txt'"
      class="hidden"
      @change="onFileChange"
    />

    <div class="flex flex-col items-center gap-3">
      <div class="w-12 h-12 rounded-xl bg-surface-600 flex items-center justify-center">
        <svg v-if="!isLoading" class="w-6 h-6 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
        </svg>
        <span v-else class="w-6 h-6 border-2 border-accent-400 border-t-transparent rounded-full animate-spin" />
      </div>
      <div>
        <p class="text-zinc-300 font-medium text-sm">{{ label ?? 'Drop file here or click to browse' }}</p>
        <p class="text-zinc-500 text-xs mt-1">Video, audio, VTT, SRT, or transcript file</p>
      </div>
    </div>
  </div>
</template>
