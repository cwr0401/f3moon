<script setup lang="ts">
import type { Combination } from '../../types/game'
import { CombType } from '../../types/game'
import TileCard from './TileCard.vue'

defineProps<{ combinations: Combination[] }>()

function combLabel(comb: Combination): string {
  if (comb.type === CombType.CombWord) return '文'
  if (comb.type === CombType.CombNumeric) return '数'
  const size = comb.tiles.length
  if (size === 3) return '坎'
  if (size === 4) return '统'
  if (size === 5) return '五统'
  return '组'
}
</script>

<template>
  <div class="flex flex-wrap gap-2 justify-center">
    <div
      v-for="(comb, idx) in combinations"
      :key="idx"
      class="flex items-center gap-0.5 bg-amber-900/10 rounded px-1 py-0.5"
    >
      <span class="text-[10px] text-amber-500 mr-0.5">{{ combLabel(comb) }}</span>
      <TileCard v-for="tile in comb.tiles" :key="tile.id" :tile="tile" />
    </div>
  </div>
</template>
