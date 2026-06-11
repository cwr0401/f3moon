<script setup lang="ts">
import { ref, computed } from 'vue'

import { useGameStore } from '../../stores/game'
import { useGameActions } from '../../composables/useGameActions'
import type { TileName } from '../../types/tile'
import { JING_TILES } from '../../types/tile'

const gameStore = useGameStore()
const actions = useGameActions()
const loading = ref(false)

const tongTile = ref<TileName>(JING_TILES[0]!)
const tongSize = ref(4)

const isMyTurn = computed(() => {
  if (!gameStore.gameState) return false
  const currentPlayerIdx = gameStore.gameState.tong_order[gameStore.gameState.tong_current]
  const myIdx = gameStore.gameState.players.findIndex(p => p?.id === gameStore.myPlayer?.id)
  return currentPlayerIdx === myIdx
})

async function handleTong() {
  loading.value = true
  try {
    await actions.doTong(tongTile.value, tongSize.value)
  } finally {
    loading.value = false
  }
}

async function handleSkip() {
  loading.value = true
  try {
    await actions.doTongSkip()
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div v-if="isMyTurn" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
    <div class="bg-bg-card rounded-xl p-6 w-full max-w-md border border-amber-900/30 shadow-2xl">
      <h2 class="text-xl font-bold text-accent mb-4">请统</h2>
      <p class="text-sm text-gray-400 mb-4">选择是否统牌。统牌可将手中4张或5张相同的牌作为一组明牌</p>

      <div class="space-y-3 mb-4">
        <div>
          <label class="block text-sm text-gray-400 mb-1">统的牌</label>
          <div class="flex gap-2">
            <button
              v-for="tile in JING_TILES"
              :key="tile"
              @click="tongTile = tile"
              :class="['px-4 py-2 rounded-lg border transition-colors', tongTile === tile ? 'bg-accent/20 border-accent text-accent' : 'bg-bg-input border-amber-900/30 text-gray-400']"
            >
              {{ tile }}
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm text-gray-400 mb-1">统牌数</label>
          <div class="flex gap-2">
            <button
              @click="tongSize = 4"
              :class="['px-4 py-2 rounded-lg border transition-colors', tongSize === 4 ? 'bg-accent/20 border-accent text-accent' : 'bg-bg-input border-amber-900/30 text-gray-400']"
            >
              4张统
            </button>
            <button
              @click="tongSize = 5"
              :class="['px-4 py-2 rounded-lg border transition-colors', tongSize === 5 ? 'bg-accent/20 border-accent text-accent' : 'bg-bg-input border-amber-900/30 text-gray-400']"
            >
              5张统
            </button>
          </div>
        </div>
      </div>

      <div class="flex gap-3">
        <button
          @click="handleSkip"
          :disabled="loading"
          class="flex-1 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-300 hover:bg-amber-900/20 transition-colors disabled:opacity-50"
        >
          不统
        </button>
        <button
          @click="handleTong"
          :disabled="loading"
          class="flex-1 py-2.5 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors disabled:opacity-50"
        >
          {{ loading ? '处理中...' : '确认统牌' }}
        </button>
      </div>
    </div>
  </div>
</template>
