<script setup lang="ts">
import { useAuthStore } from './stores/auth'
import AppHeader from './components/common/AppHeader.vue'
import Toast from './components/common/Toast.vue'

const auth = useAuthStore()
auth.init()
</script>

<template>
  <div class="min-h-screen bg-bg-dark text-gray-100">
    <AppHeader v-if="auth.isLoggedIn" />
    <main class="container mx-auto px-4 py-6">
      <RouterView />
    </main>
    <Toast />
  </div>
</template>
