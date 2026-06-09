<script setup lang="ts">
import type { Room } from '../../types/room'
import { RoomStatus } from '../../types/room'
import { GameMode } from '../../types/game'

defineProps<{ room: Room }>()

function modeLabel(mode: GameMode) {
  return mode === GameMode.Mode3Player ? '3人定庄' : '4人含歇家'
}

function statusLabel(status: RoomStatus) {
  switch (status) {
    case RoomStatus.RoomWaiting: return '等待中'
    case RoomStatus.RoomPlaying: return '游戏中'
    case RoomStatus.RoomFinished: return '已结束'
  }
}
</script>

<template>
  <div class="bg-bg-card rounded-xl p-4 border border-amber-900/30">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-bold text-accent">{{ room.name }}</h2>
      <span
        :class="[
          'px-3 py-1 rounded text-sm',
          room.status === RoomStatus.RoomWaiting ? 'bg-green-900/40 text-green-300' : 'bg-amber-900/40 text-amber-300'
        ]"
      >
        {{ statusLabel(room.status) }}
      </span>
    </div>
    <div class="flex gap-4 text-sm text-gray-400 mt-2">
      <span>模式: {{ modeLabel(room.mode) }}</span>
      <span>•</span>
      <span>房间ID: {{ room.id.slice(0, 8) }}</span>
    </div>
  </div>
</template>
