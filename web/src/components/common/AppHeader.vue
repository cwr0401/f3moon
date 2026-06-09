<script setup lang="ts">
import { useAuthStore } from '../../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <header class="bg-bg-card border-b border-amber-900/30 px-6 py-3 flex items-center justify-between">
    <router-link to="/lobby" class="flex items-center gap-2 text-accent font-bold text-lg hover:opacity-80 transition-opacity">
      <span class="text-2xl">🀄</span>
      <span>荆楚花牌</span>
    </router-link>
    <div class="flex items-center gap-4">
      <span class="text-gray-300 text-sm">{{ auth.user?.nickname || auth.user?.email }}</span>
      <button
        @click="handleLogout"
        class="px-3 py-1 text-sm rounded bg-bg-input hover:bg-amber-900/40 transition-colors"
      >
        退出
      </button>
    </div>
  </header>
</template>
