<script setup lang="ts">
import type { RoomPlayer } from '../../types/room'

defineProps<{
  player: RoomPlayer | null
  seatIndex: number
  maxPlayers: number
}>()

const seatLabels = ['庄家', '闲一', '闲二', '歇家']
</script>

<template>
  <div
    :class="[
      'rounded-xl p-4 border transition-colors min-h-[100px] flex flex-col items-center justify-center',
      player ? 'bg-bg-card border-amber-900/40' : 'bg-bg-dark/50 border-amber-900/10 border-dashed'
    ]"
  >
    <div v-if="seatIndex < maxPlayers">
      <template v-if="player">
        <div class="text-lg font-bold" :class="player.is_ai ? 'text-blue-300' : 'text-gray-100'">
          {{ player.is_ai ? '🤖' : '👤' }} {{ player.name }}
        </div>
        <div class="text-sm text-gray-400 mt-1">{{ seatLabels[seatIndex] }}</div>
        <div v-if="player.ready" class="mt-2 text-xs text-green-400 bg-green-900/20 px-2 py-0.5 rounded">已准备</div>
      </template>
      <template v-else>
        <div class="text-gray-600 text-sm">{{ seatLabels[seatIndex] }}</div>
        <div class="text-gray-700 text-xs mt-1">等待加入</div>
      </template>
    </div>
    <div v-else class="text-gray-700 text-sm">不启用</div>
  </div>
</template>
