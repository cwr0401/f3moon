<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'
import { useGameActions } from '../../composables/useGameActions'
import { GamePhase } from '../../types/game'

const gameStore = useGameStore()
const actions = useGameActions()

const showDeal = computed(() =>
  gameStore.currentPhase === GamePhase.PhaseDeal
)

const showDraw = computed(() =>
  gameStore.currentPhase === GamePhase.PhasePlay && gameStore.isMyTurn && !gameStore.gameState?.last_discard
)

const showDiscard = computed(() =>
  gameStore.currentPhase === GamePhase.PhasePlay && gameStore.isMyTurn && gameStore.selectedTileId !== null
)

const showWin = computed(() =>
  gameStore.currentPhase === GamePhase.PhasePlay && gameStore.isMyTurn
)

const showGanta = computed(() =>
  gameStore.currentPhase === GamePhase.PhasePlay && gameStore.isMyTurn && gameStore.gameState?.last_discard !== null
)

const showPass = computed(() =>
  gameStore.currentPhase === GamePhase.PhasePlay && !gameStore.isMyTurn && gameStore.gameState?.last_discard !== null
)
</script>

<template>
  <div class="flex gap-2 justify-center py-2">
    <button
      v-if="showDeal"
      @click="actions.doDeal()"
      class="px-4 py-2 rounded-lg bg-blue-600 text-white font-bold hover:bg-blue-700 transition-colors"
    >
      发牌
    </button>
    <button
      v-if="showDraw"
      @click="actions.doDraw()"
      class="px-4 py-2 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors"
    >
      起牌
    </button>
    <button
      v-if="showDiscard"
      @click="gameStore.selectedTileId && actions.doDiscard(gameStore.selectedTileId)"
      class="px-4 py-2 rounded-lg bg-red-600 text-white font-bold hover:bg-red-700 transition-colors"
    >
      出牌
    </button>
    <button
      v-if="showGanta"
      @click="actions.doGanta()"
      class="px-4 py-2 rounded-lg bg-purple-600 text-white hover:bg-purple-700 transition-colors"
    >
      赶塔
    </button>
    <button
      v-if="showWin"
      @click="actions.doWin()"
      class="px-4 py-2 rounded-lg bg-green-600 text-white font-bold hover:bg-green-700 transition-colors"
    >
      和牌
    </button>
    <button
      v-if="showPass"
      @click="actions.doPass()"
      class="px-4 py-2 rounded-lg bg-bg-input border border-amber-900/30 text-gray-300 hover:bg-amber-900/20 transition-colors"
    >
      过
    </button>
  </div>
</template>
