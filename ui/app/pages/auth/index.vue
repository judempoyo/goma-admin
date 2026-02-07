<script setup lang="ts">

definePageMeta({
  layout: 'guest',
  middleware: ['guest'],
})

const router = useRouter()

onMounted(() => {
  router.replace('/auth/login')
   
})
</script>

<template>
  <div class="h-screen flex items-center justify-center">
    <div class="text-gray-700 text-center">
      <p class="text-base font-medium">Redirection vers la page de connexion...</p>
    </div>
  </div>
</template>
