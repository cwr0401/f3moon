<script setup lang="ts">
import { ref } from 'vue'
import { useGameActions } from '../../composables/useGameActions'
import { JING_TILES } from '../../types/tile'
import type { TileName } from '../../types/tile'

const actions = useGameActions()
const loading = ref(false)
const selected = ref<TileName>(JING_TILES[0]!)

async function handleSelect() {
  loading.value = true
  try {
    await actions.doDangJing(selected.value)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
    <div class="bg-bg-card rounded-xl p-6 w-full max-w-sm border border-amber-900/30 shadow-2xl">
      <h2 class="text-xl font-bold text-accent mb-4">选择当经</h2>
      <p class="text-sm text-gray-400 mb-4">选择一张经牌作为你的当经</p>

      <div class="flex gap-3 mb-4 justify-center">
        <button
          v-for="tile in JING_TILES"
          :key="tile"
          @click="selected = tile"
          :class="[
            'w-20 h-24 rounded-lg border-2 flex items-center justify-center text-3xl font-bold transition-all',
            selected === tile
              ? 'border-accent bg-accent/20 text-accent -translate-y-1'
              : 'border-amber-900/30 bg-bg-input text-gray-300 hover:border-amber-700'
          ]"
        >
          {{ tile }}
        </button>
      </div>

      <button
        @click="handleSelect"
        :disabled="loading"
        class="w-full py-3 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors disabled:opacity-50"
      >
        {{ loading ? '选择中...' : '确认当经' }}
      </button>
    </div>
  </div>
</template>
