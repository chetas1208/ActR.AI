<template>
  <div class="min-h-screen flex flex-col">
    <header class="sticky top-0 z-50 border-b border-zinc-800/50 backdrop-blur-md bg-surface-900/80">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-14">
          <NuxtLink to="/" class="flex items-center gap-2.5 group">
            <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-accent-400 to-accent-500 flex items-center justify-center shadow-lg shadow-accent-500/25 group-hover:shadow-accent-500/40 transition-shadow">
              <svg class="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
              </svg>
            </div>
            <span class="font-semibold text-zinc-100 tracking-tight">Gorube <span class="text-accent-400">Flow</span></span>
          </NuxtLink>

          <nav class="hidden sm:flex items-center gap-1">
            <NuxtLink to="/" class="btn-ghost text-sm">Home</NuxtLink>
            <NuxtLink to="/new" class="btn-ghost text-sm">New Workflow</NuxtLink>
          </nav>

          <NuxtLink to="/new" class="btn-primary text-sm py-2 px-4">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            New
          </NuxtLink>
        </div>
      </div>
    </header>

    <main class="flex-1">
      <slot />
    </main>

    <footer class="border-t border-zinc-800/50 py-6">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-3">
        <p class="text-zinc-500 text-sm">Built on <span class="text-zinc-400">GoTube</span> · Powered by <span class="text-accent-400">Gorube Flow</span></p>
        <div class="flex items-center gap-4">
          <SponsorBadge name="Tigris" color="blue" />
          <SponsorBadge name="Daytona" color="purple" />
          <SponsorBadge name="InsForge" color="orange" />
          <SponsorBadge name="Rtrvr" color="pink" />
          <SponsorBadge name="Vercel" color="indigo" />
        </div>
      </div>
    </footer>
  </div>
</template>
