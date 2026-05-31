<script setup lang="ts">
defineProps<{ message: string; title?: string }>()
</script>

<template>
  <div class="rounded-xl p-4 border border-red-500/20 bg-red-500/5 flex items-start gap-3">
    <svg class="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
      <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
    </svg>
    <div>
      <p v-if="title" class="font-medium text-red-300 text-xs mb-0.5">{{ title }}</p>
      <p class="text-zinc-400 text-xs">{{ message }}</p>
    </div>
  </div>
</template>
