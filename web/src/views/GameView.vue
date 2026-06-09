<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useGameStore } from '../stores/game'
import { useWebSocket } from '../composables/useWebSocket'
import { GamePhase } from '../types/game'
import HandTiles from '../components/game/HandTiles.vue'
import OpenCombs from '../components/game/OpenCombs.vue'
import DiscardPile from '../components/game/DiscardPile.vue'
import PlayerSeat from '../components/game/PlayerSeat.vue'
import ActionBar from '../components/game/ActionBar.vue'
import CutSelector from '../components/game/CutSelector.vue'
import TongDialog from '../components/game/TongDialog.vue'
import DangJingDialog from '../components/game/DangJingDialog.vue'
import GameResult from '../components/game/GameResult.vue'

const route = useRoute()
const gameStore = useGameStore()
const gameId = ref(route.params.id as string)
const { status: wsStatus } = useWebSocket(gameId)

onMounted(async () => {
  await gameStore.fetchGame(gameId.value)
})

onUnmounted(() => {
  gameStore.reset()
})
</script>

<template>
  <div class="h-[calc(100vh-64px)] flex flex-col">
    <!-- WS Status indicator -->
    <div v-if="wsStatus !== 'open'" class="text-center py-1 text-xs bg-red-900/40 text-red-300">
      {{ wsStatus === 'connecting' ? '连接中...' : '连接断开，正在重连...' }}
    </div>

    <div v-if="gameStore.gameState" class="flex-1 game-table p-4 flex flex-col">
      <!-- Opponent seats (top) -->
      <div class="flex justify-around mb-4">
        <PlayerSeat
          v-for="i in [2, 1]"
          :key="i"
          :player="gameStore.gameState.players[i] ?? null"
          :is-current="gameStore.gameState.current_turn === i"
          :my-id="gameStore.myPlayer?.id"
        />
      </div>

      <!-- Middle: discard pile -->
      <div class="flex-1 flex items-center justify-center">
        <DiscardPile :tiles="gameStore.gameState.discard_pile" :last-tile="gameStore.gameState.last_discard" />
      </div>

      <!-- My area (bottom) -->
      <div class="space-y-2">
        <!-- My open combinations -->
        <OpenCombs :combinations="gameStore.myOpenCombs" />

        <!-- My hand -->
        <HandTiles
          :tiles="gameStore.myHand"
          :selected-id="gameStore.selectedTileId"
          @select="gameStore.selectTile"
        />

        <!-- Action bar -->
        <ActionBar />
      </div>
    </div>

    <!-- Phase dialogs -->
    <CutSelector v-if="gameStore.currentPhase === GamePhase.PhaseCut && gameStore.isCutPlayer" />
    <TongDialog v-if="gameStore.currentPhase === GamePhase.PhaseTongAsk" />
    <DangJingDialog v-if="gameStore.currentPhase === GamePhase.PhasePlay && !gameStore.myPlayer?.dang_jing" />
    <GameResult v-if="gameStore.currentPhase === GamePhase.PhaseFinished" />
  </div>
</template>
