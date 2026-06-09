<script setup lang="ts">
import type { Tile } from '../../types/tile'
import TileCard from './TileCard.vue'

defineProps<{
  tiles: Tile[]
  selectedId: number | null
}>()

const emit = defineEmits<{ select: [tileId: number] }>()
</script>

<template>
  <div class="flex flex-wrap gap-1 justify-center">
    <TileCard
      v-for="tile in tiles"
      :key="tile.id"
      :tile="tile"
      :selected="tile.id === selectedId"
      @click="emit('select', tile.id)"
    />
  </div>
</template>
