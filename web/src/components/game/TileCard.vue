<script setup lang="ts">
import type { Tile } from '../../types/tile'
import { RED_TILE_NAMES, isJingName } from '../../types/tile'

const props = defineProps<{
  tile: Tile
  selected?: boolean
  faceDown?: boolean
}>()

const emit = defineEmits<{ click: [tileId: number] }>()

const isRed = RED_TILE_NAMES.has(props.tile.name) || props.tile.name === '别'
const isJing = isJingName(props.tile.name)
</script>

<template>
  <div
    v-if="faceDown"
    class="tile-card tile-back"
  >
    <span class="text-slate-500 text-xs">花</span>
  </div>
  <div
    v-else
    :class="[
      'tile-card',
      { 'tile-red': isRed, 'tile-black': !isRed, 'tile-jing': isJing, 'selected': selected }
    ]"
    @click="emit('click', tile.id)"
  >
    <span v-if="tile.is_flower" class="tile-flower">花</span>
    <span class="tile-name">{{ tile.name }}</span>
    <span v-if="isRed" class="tile-red-dot"></span>
  </div>
</template>
