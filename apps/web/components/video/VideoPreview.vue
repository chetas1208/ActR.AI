<script setup lang="ts">
import type { Video } from '~/types'

defineProps<{ video: Video; playbackUrl?: string }>()
</script>

<template>
  <div class="card overflow-hidden">
    <!-- YouTube embed -->
    <template v-if="video.youtubeEmbedUrl">
      <div class="aspect-video bg-black">
        <iframe
          :src="video.youtubeEmbedUrl"
          class="w-full h-full"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowfullscreen
          :title="video.title"
        />
      </div>
    </template>

    <!-- Uploaded/direct file video player -->
    <template v-else-if="playbackUrl">
      <div class="aspect-video bg-black">
        <video
          :src="playbackUrl"
          controls
          class="w-full h-full"
          :poster="video.thumbnailUrl ?? undefined"
        >
          Your browser does not support video playback.
        </video>
      </div>
    </template>

    <!-- Thumbnail only fallback -->
    <template v-else-if="video.thumbnailUrl">
      <div class="aspect-video bg-surface-700 relative">
        <img :src="video.thumbnailUrl" :alt="video.title" class="w-full h-full object-cover" />
        <div class="absolute inset-0 flex items-center justify-center">
          <div class="w-14 h-14 rounded-full bg-black/60 flex items-center justify-center">
            <svg class="w-6 h-6 text-white ml-1" fill="currentColor" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z" />
            </svg>
          </div>
        </div>
      </div>
    </template>

    <!-- Metadata -->
    <div class="p-4">
      <h2 class="font-semibold text-zinc-100 text-sm leading-tight">{{ video.title }}</h2>
      <div class="flex items-center gap-3 mt-2 text-xs text-zinc-500">
        <span v-if="video.durationSeconds">{{ formatDuration(video.durationSeconds) }}</span>
        <span :class="['badge text-xs', video.status === 'ready' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-surface-600 text-zinc-400']">
          {{ video.status }}
        </span>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
function formatDuration(s: number): string {
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
  return `${m}:${String(sec).padStart(2, '0')}`
}
export default { methods: { formatDuration } }
</script>
