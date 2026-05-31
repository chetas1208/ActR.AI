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
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:3001',
    },
  },
  app: {
    head: {
      title: 'Gorube Flow – Video to Agent Workflow',
      meta: [
        { name: 'description', content: 'Turn videos into agent workflows. Extract actions, verify claims, execute code safely.' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap' },
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
              900: '#0a0a0f',
              800: '#111118',
              700: '#1a1a24',
              600: '#22222e',
              500: '#2c2c3a',
            },
            accent: {
              500: '#6366f1',
              400: '#818cf8',
              300: '#a5b4fc',
            },
            status: {
              pending: '#f59e0b',
              running: '#3b82f6',
              completed: '#10b981',
              failed: '#ef4444',
              waiting: '#8b5cf6',
            },
          },
          animation: {
            'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
            'fade-in': 'fadeIn 0.3s ease-out',
            'slide-up': 'slideUp 0.3s ease-out',
          },
          keyframes: {
            fadeIn: { from: { opacity: '0' }, to: { opacity: '1' } },
            slideUp: { from: { transform: 'translateY(8px)', opacity: '0' }, to: { transform: 'translateY(0)', opacity: '1' } },
          },
        },
      },
    },
  },
})
