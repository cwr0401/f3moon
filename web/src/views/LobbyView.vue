<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useZoneStore } from '../stores/zone'
import type { GameZone } from '../types/zone'

import RoomCard from '../components/lobby/RoomCard.vue'
import CreateRoomModal from '../components/lobby/CreateRoomModal.vue'

const router = useRouter()
const roomStore = useRoomStore()
const zoneStore = useZoneStore()
const showCreate = ref(false)
const selectedZoneId = ref<string | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await zoneStore.fetchZones()
  // 选择第一个游戏区作为默认
  if (zoneStore.zones.length > 0) {
    const firstZone = zoneStore.zones[0] as GameZone | undefined
    if (firstZone) {
      selectedZoneId.value = firstZone.id
    }
  }
  await roomStore.fetchRooms(selectedZoneId.value ?? undefined)
  pollTimer = setInterval(() => roomStore.fetchRooms(selectedZoneId.value ?? undefined), 5000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

async function handleZoneChange(zone: GameZone) {
  selectedZoneId.value = zone.id
  await roomStore.fetchRooms(zone.id)
}

async function handleJoin(roomId: string) {
  await roomStore.joinRoom(roomId)
  router.push(`/room/${roomId}`)
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-2xl font-bold text-accent">游戏大厅</h2>
      <button
        @click="showCreate = true"
        class="px-5 py-2 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors"
      >
        创建房间
      </button>
    </div>

    <!-- 游戏区选择 -->
    <div v-if="zoneStore.zones.length > 0" class="mb-6">
      <div class="flex gap-3 overflow-x-auto pb-2">
        <button
          v-for="zone in zoneStore.zones"
          :key="zone.id"
          @click="handleZoneChange(zone)"
          :class="[
            'flex-shrink-0 px-4 py-2 rounded-lg border transition-colors',
            selectedZoneId === zone.id ? 'bg-accent/20 border-accent text-accent' : 'bg-bg-card border-amber-900/30 text-gray-400 hover:bg-amber-900/20'
          ]"
        >
          <div class="font-medium">{{ zone.name }}</div>
          <div class="text-xs opacity-70">{{ zone.description }} ({{ zone.room_count }}/{{ zone.max_rooms }})</div>
        </button>
      </div>
    </div>

    <div v-if="roomStore.rooms.length === 0" class="text-center py-20 text-gray-500">
      暂无房间，点击"创建房间"开始游戏
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <RoomCard
        v-for="room in roomStore.rooms"
        :key="room.id"
        :room="room"
        @join="handleJoin"
      />
    </div>

    <CreateRoomModal
      v-if="showCreate"
      :zone-id="selectedZoneId"
      @close="showCreate = false"
    />
  </div>
</template>
