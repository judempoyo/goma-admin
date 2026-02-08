import Aura from '@primeuix/themes/aura';
import tailwindcss from "@tailwindcss/vite";


export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      appUrl: '', 
      apiUrl: '',
      appName: '',
      appVersion:''
    }
  },
  modules: [
    '@primevue/nuxt-module',
    '@pinia/nuxt',
    '@nuxtjs/color-mode'
  ],
  primevue: {
    options: {
      theme: {
        preset: Aura,
        options: {
          prefix: 'p',
          darkModeSelector: '.dark',
        },
        components: {
          exclude: []
        }
      },
      
    }
  },
  css: [
    "./app/assets/css/main.css",
    'primeicons/primeicons.css'
  ],
  vite: {
    plugins: [
      tailwindcss(),
    ],
  },
  nitro: {
    routeRules: {
      '/_nuxt/**': { cors: true }
    }
  },
  pinia: {
    storesDirs: ['./stores/**'],
  },
  colorMode: {
    preference: 'system',
    fallback: 'light', 
  },
  app: {
    baseURL: '/',
  },
})