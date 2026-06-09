<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useAuthStore } from '../stores/auth'
import { RoomStatus } from '../types/room'
import SeatSlot from '../components/room/SeatSlot.vue'
import RoomInfo from '../components/room/RoomInfo.vue'

const route = useRoute()
const router = useRouter()
const roomStore = useRoomStore()
const auth = useAuthStore()
const roomId = route.params.id as string

let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await roomStore.fetchRoom(roomId)
  pollTimer = setInterval(() => roomStore.fetchRoom(roomId), 3000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

async function handleAddAI() {
  await roomStore.addAI(roomId)
  await roomStore.fetchRoom(roomId)
}

async function handleLeave() {
  await roomStore.leaveRoom(roomId)
  roomStore.clearCurrentRoom()
  router.push('/lobby')
}

async function handleStart() {
  try {
    await roomStore.startGame(roomId)
    router.push(`/game/${roomId}`)
  } catch {
    // Error handled by store
  }
}

async function handleReady() {
  if (!roomStore.currentRoom) return
  const me = findMyPlayer()
  if (!me) return
  const newReady = !me.ready
  await roomStore.ready(roomId, newReady)
  await roomStore.fetchRoom(roomId)
}

function findMyPlayer() {
  if (!roomStore.currentRoom) return null
  return roomStore.currentRoom.players.find(p => p?.id === auth.user?.id) ?? null
}

function isOwner(): boolean {
  return auth.user?.id === roomStore.currentRoom?.owner
}

function amIReady(): boolean {
  return findMyPlayer()?.ready ?? false
}
</script>

<template>
  <div v-if="roomStore.currentRoom" class="max-w-2xl mx-auto">
    <RoomInfo :room="roomStore.currentRoom" />

    <div class="grid grid-cols-2 gap-4 mt-6">
      <SeatSlot
        v-for="(player, idx) in roomStore.currentRoom.players"
        :key="idx"
        :player="player"
        :seat-index="idx"
        :max-players="roomStore.currentRoom.max_players"
      />
    </div>

    <div class="flex gap-3 mt-8 justify-center">
      <button
        v-if="isOwner() && roomStore.currentRoom.status === RoomStatus.RoomWaiting"
        @click="handleAddAI"
        class="px-5 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 hover:bg-amber-900/30 transition-colors"
      >
        添加 AI
      </button>
      <button
        v-if="roomStore.currentRoom.status === RoomStatus.RoomWaiting"
        @click="handleReady"
        :class="[
          'px-5 py-2.5 rounded-lg font-bold transition-colors',
          amIReady()
            ? 'bg-green-900/40 border border-green-700/40 text-green-300 hover:bg-green-900/60'
            : 'bg-accent text-bg-dark hover:bg-accent-dim'
        ]"
      >
        {{ amIReady() ? '取消准备' : '准备' }}
      </button>
      <button
        v-if="isOwner() && roomStore.currentRoom.status === RoomStatus.RoomWaiting"
        @click="handleStart"
        :disabled="!roomStore.currentRoom.players.every(p => p === null || p.ready) || roomStore.currentRoom.players.filter(p => p !== null).length < roomStore.currentRoom.max_players"
        class="px-5 py-2.5 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors disabled:opacity-50"
      >
        开始游戏
      </button>
      <button
        @click="handleLeave"
        class="px-5 py-2.5 rounded-lg bg-red-900/40 border border-red-700/40 hover:bg-red-900/60 transition-colors"
      >
        离开房间
      </button>
    </div>
  </div>
</template>
