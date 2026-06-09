<script setup lang="ts">
import type { Room } from '../../types/room'
import { RoomStatus } from '../../types/room'
import { GameMode } from '../../types/game'

defineProps<{ room: Room }>()
const emit = defineEmits<{ join: [roomId: string] }>()

function modeLabel(mode: GameMode) {
  return mode === GameMode.Mode3Player ? '3人' : '4人'
}

function statusLabel(status: RoomStatus) {
  switch (status) {
    case RoomStatus.RoomWaiting: return '等待中'
    case RoomStatus.RoomPlaying: return '游戏中'
    case RoomStatus.RoomFinished: return '已结束'
  }
}

function playerCount(room: Room): number {
  return room.players.filter(p => p !== null).length
}
</script>

<template>
  <div class="bg-bg-card rounded-xl p-5 border border-amber-900/30 hover:border-accent/50 transition-colors">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-lg font-bold text-gray-100">{{ room.name }}</h3>
      <span
        :class="[
          'px-2 py-0.5 rounded text-xs font-medium',
          room.status === RoomStatus.RoomWaiting ? 'bg-green-900/40 text-green-300' : 'bg-amber-900/40 text-amber-300'
        ]"
      >
        {{ statusLabel(room.status) }}
      </span>
    </div>
    <div class="flex items-center gap-3 text-sm text-gray-400 mb-4">
      <span>{{ modeLabel(room.mode) }}模式</span>
      <span>•</span>
      <span>{{ playerCount(room) }}/{{ room.max_players }}人</span>
    </div>
    <button
      @click="emit('join', room.id)"
      :disabled="room.status !== RoomStatus.RoomWaiting || playerCount(room) >= room.max_players"
      class="w-full py-2 rounded-lg bg-accent/20 text-accent hover:bg-accent/30 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    >
      {{ room.status === RoomStatus.RoomWaiting && playerCount(room) < room.max_players ? '加入' : '已满' }}
    </button>
  </div>
</template>
