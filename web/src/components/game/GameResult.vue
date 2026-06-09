<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'
import { useAuthStore } from '../../stores/auth'
import { WinType } from '../../types/game'
import { useRouter } from 'vue-router'

const gameStore = useGameStore()
const auth = useAuthStore()
const router = useRouter()

const winner = computed(() => {
  const gs = gameStore.gameState
  if (!gs || gs.winner < 0) return null
  return gs.players[gs.winner]
})

const isMyWin = computed(() => winner.value?.id === auth.user?.id)

const winTypeLabel = computed(() => {
  switch (gameStore.gameState?.win_type) {
    case WinType.WinZiMo: return '自摸'
    case WinType.WinDianPao: return '点炮'
    case WinType.WinTianHu: return '天胡'
    case WinType.WinHaiDi: return '海底捞月'
    default: return ''
  }
})

const isHuang = computed(() => gameStore.gameState?.winner === -1 && gameStore.gameState?.phase === 7)

function backToLobby() {
  gameStore.reset()
  router.push('/lobby')
}
</script>

<template>
  <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
    <div class="bg-bg-card rounded-2xl p-8 w-full max-w-md border border-amber-900/30 shadow-2xl text-center">
      <template v-if="isHuang">
        <div class="text-5xl mb-4">🏜️</div>
        <h2 class="text-2xl font-bold text-gray-300 mb-2">荒牌</h2>
        <p class="text-gray-500">牌局无人和牌</p>
      </template>
      <template v-else-if="winner">
        <div class="text-5xl mb-4">{{ isMyWin ? '🎉' : '😔' }}</div>
        <h2 :class="['text-2xl font-bold mb-2', isMyWin ? 'text-accent' : 'text-gray-300']">
          {{ isMyWin ? '恭喜和牌!' : `${winner.name} 和牌` }}
        </h2>
        <p class="text-gray-400">{{ winTypeLabel }}</p>
      </template>

      <button
        @click="backToLobby"
        class="mt-6 px-8 py-3 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors"
      >
        返回大厅
      </button>
    </div>
  </div>
</template>
