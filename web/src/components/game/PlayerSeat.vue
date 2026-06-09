<script setup lang="ts">
import type { Player } from '../../types/game'
import { PlayerRole } from '../../types/game'

defineProps<{
  player: Player | null
  isCurrent: boolean
  myId?: string
}>()

const roleLabels: Record<number, string> = {
  [PlayerRole.RoleDealer]: '庄',
  [PlayerRole.RoleIdle1]: '闲一',
  [PlayerRole.RoleIdle2]: '闲二',
  [PlayerRole.RoleRest]: '歇家',
}
</script>

<template>
  <div
    v-if="player"
    :class="[
      'flex flex-col items-center gap-1 px-4 py-2 rounded-lg transition-all',
      isCurrent ? 'bg-accent/20 ring-2 ring-accent/50' : 'bg-bg-card/50'
    ]"
  >
    <div class="text-sm font-bold" :class="isCurrent ? 'text-accent' : 'text-gray-300'">
      {{ player.name }}
      <span v-if="player.id === myId" class="text-xs text-green-400">(你)</span>
    </div>
    <div class="text-xs text-gray-500">
      {{ roleLabels[player.role] }}
      <span v-if="player.dang_jing" class="ml-1 text-jing">经:{{ player.dang_jing }}</span>
    </div>
    <div class="text-xs text-gray-500">
      手牌: {{ player.hand.length }} | 对: {{ player.pair_count }}/2
    </div>
    <div v-if="!player.online" class="text-xs text-red-400">离线</div>
  </div>
</template>
