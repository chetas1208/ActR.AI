// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },
  modules: [
    '@nuxtjs/tailwindcss',
    '@pinia/nuxt',
  ],
  runtimeConfig: {
    public: {
      appName: process.env.NUXT_PUBLIC_APP_NAME || 'ActR.AI',
      appUrl: process.env.NUXT_PUBLIC_APP_URL || 'http://localhost:3000',
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || (process.env.NODE_ENV === 'production' ? '' : 'http://localhost:8080'),
      enableYoutubeInput: process.env.NUXT_PUBLIC_ENABLE_YOUTUBE_INPUT !== 'false',
      enableUploadInput: process.env.NUXT_PUBLIC_ENABLE_UPLOAD_INPUT !== 'false',
      enableDirectFileInput: process.env.NUXT_PUBLIC_ENABLE_DIRECT_FILE_INPUT !== 'false',
      enableRtrvrResearch: process.env.NUXT_PUBLIC_ENABLE_RTRVR_RESEARCH !== 'false',
      enableDaytonaExecution: process.env.NUXT_PUBLIC_ENABLE_DAYTONA_EXECUTION !== 'false',
      enableNvidiaAI: process.env.NUXT_PUBLIC_ENABLE_NVIDIA_AI !== 'false',
      workflowPollIntervalMs: parseInt(process.env.NUXT_PUBLIC_WORKFLOW_POLL_INTERVAL_MS || '2000'),
      workflowAutoAdvance: process.env.NUXT_PUBLIC_WORKFLOW_AUTO_ADVANCE !== 'false',
    },
  },
  app: {
    head: {
      title: 'ActR.AI – Video to Agent Workflow',
      meta: [
        { name: 'description', content: 'ActR.AI turns video into action. Extract, research, and execute — powered by NVIDIA NIM.' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { property: 'og:title', content: 'ActR.AI' },
        { property: 'og:description', content: 'Turn video into agent workflows.' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap',
        },
      ],
    },
  },
  css: ['~/assets/css/globals.css'],
  tailwindcss: {
    config: {
      darkMode: 'class',
      theme: {
        extend: {
          fontFamily: {
            sans: ['Inter', 'system-ui', 'sans-serif'],
            mono: ['JetBrains Mono', 'monospace'],
          },
          colors: {
            surface: {
              950: '#070711',
              900: '#0c0c18',
              800: '#111124',
              700: '#18182e',
              600: '#1f1f38',
              500: '#272744',
            },
            cyan: {
              400: '#22d3ee',
              500: '#06b6d4',
            },
            violet: {
              400: '#a78bfa',
              500: '#8b5cf6',
            },
            electric: {
              400: '#60a5fa',
              500: '#3b82f6',
            },
            status: {
              pending: '#f59e0b',
              running: '#3b82f6',
              completed: '#10b981',
              failed: '#ef4444',
              waiting: '#8b5cf6',
              researching: '#22d3ee',
            },
          },
          backgroundImage: {
            'grid-pattern': 'linear-gradient(rgba(34,211,238,0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(34,211,238,0.03) 1px, transparent 1px)',
          },
          backgroundSize: {
            'grid': '40px 40px',
          },
          animation: {
            'pulse-slow': 'pulse 3s cubic-bezier(0.4,0,0.6,1) infinite',
            'fade-in': 'fadeIn 0.35s ease-out',
            'slide-up': 'slideUp 0.35s ease-out',
            'glow': 'glow 2s ease-in-out infinite alternate',
          },
          keyframes: {
            fadeIn:  { from: { opacity: '0' }, to: { opacity: '1' } },
            slideUp: { from: { transform: 'translateY(10px)', opacity: '0' }, to: { transform: 'translateY(0)', opacity: '1' } },
            glow:    { from: { boxShadow: '0 0 10px rgba(34,211,238,0.2)' }, to: { boxShadow: '0 0 24px rgba(34,211,238,0.5)' } },
          },
          boxShadow: {
            'glow-cyan': '0 0 20px rgba(34,211,238,0.3)',
            'glow-violet': '0 0 20px rgba(139,92,246,0.3)',
          },
        },
      },
    },
  },
})
