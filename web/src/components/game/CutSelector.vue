<script setup lang="ts">
import { ref } from 'vue'
import { useGameActions } from '../../composables/useGameActions'

const actions = useGameActions()
const position = ref(73)
const loading = ref(false)

async function handleCut() {
  loading.value = true
  try {
    await actions.doCut(position.value)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
    <div class="bg-bg-card rounded-xl p-6 w-full max-w-md border border-amber-900/30 shadow-2xl">
      <h2 class="text-xl font-bold text-accent mb-4">切牌 (腰牌)</h2>
      <p class="text-sm text-gray-400 mb-4">选择切牌位置，公牌将从该位置一分为二</p>

      <div class="mb-4">
        <input
          v-model.number="position"
          type="range"
          min="37"
          max="110"
          class="w-full accent-amber-500"
        />
        <div class="flex justify-between text-xs text-gray-500 mt-1">
          <span>37</span>
          <span class="text-accent font-bold text-sm">{{ position }}</span>
          <span>110</span>
        </div>
      </div>

      <button
        @click="handleCut"
        :disabled="loading"
        class="w-full py-3 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors disabled:opacity-50"
      >
        {{ loading ? '切牌中...' : '确认切牌' }}
      </button>
    </div>
  </div>
</template>
