<script setup lang="ts">
import type { Tile } from '../../types/tile'
import TileCard from './TileCard.vue'

defineProps<{
  tiles: Tile[]
  lastTile: Tile | null
}>()
</script>

<template>
  <div class="flex flex-wrap gap-0.5 max-w-xl justify-center p-3">
    <TileCard
      v-for="tile in tiles"
      :key="tile.id"
      :tile="tile"
      :class="{ 'ring-2 ring-amber-400': lastTile && tile.id === lastTile.id }"
    />
    <div v-if="tiles.length === 0" class="text-gray-600 text-sm py-4">出牌区</div>
  </div>
</template>
