<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'

import RoomCard from '../components/lobby/RoomCard.vue'
import CreateRoomModal from '../components/lobby/CreateRoomModal.vue'

const router = useRouter()
const roomStore = useRoomStore()
const showCreate = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await roomStore.fetchRooms()
  pollTimer = setInterval(() => roomStore.fetchRooms(), 5000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

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
      @close="showCreate = false"
    />
  </div>
</template>
