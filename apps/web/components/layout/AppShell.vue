<template>
  <div class="min-h-screen flex flex-col">
    <!-- Nav -->
    <header class="sticky top-0 z-50 border-b border-zinc-800/50 backdrop-blur-xl bg-surface-950/80">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-14">
          <!-- Logo -->
          <NuxtLink to="/" class="flex items-center gap-2.5 group">
            <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-cyan-400 to-violet-500 flex items-center justify-center shadow-glow-cyan group-hover:shadow-[0_0_28px_rgba(34,211,238,0.5)] transition-all duration-300">
              <svg class="w-3.5 h-3.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
              </svg>
            </div>
            <span class="font-bold text-sm tracking-tight">
              <span class="text-zinc-100">ActR</span><span class="text-cyan-400">.</span><span class="text-zinc-100">AI</span>
            </span>
          </NuxtLink>

          <!-- Nav links -->
          <nav class="hidden sm:flex items-center gap-1">
            <NuxtLink to="/" class="btn-ghost text-sm">Home</NuxtLink>
            <NuxtLink to="/new" class="btn-ghost text-sm">New Workflow</NuxtLink>
          </nav>

          <NuxtLink to="/new" class="btn-primary text-sm py-2 px-4">
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            Start
          </NuxtLink>
        </div>
      </div>
    </header>

    <main class="flex-1">
      <slot />
    </main>

    <!-- Footer -->
    <footer class="border-t border-zinc-800/50 py-5 mt-auto">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-3">
        <p class="text-zinc-600 text-xs">
          <span class="text-zinc-400 font-medium">ActR.AI</span> — Video to agent workflow engine
        </p>
        <div class="flex items-center gap-2.5">
          <ProviderChip name="NVIDIA NIM" />
          <ProviderChip name="Tigris" />
          <ProviderChip name="Daytona" />
          <ProviderChip name="Rtrvr" />
          <ProviderChip name="InsForge" />
          <ProviderChip name="Vercel" />
        </div>
      </div>
    </footer>
  </div>
</template>
