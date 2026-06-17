import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '../composables/useApi'
import type { Room, RoomPlayer } from '../types/room'
import { GameMode } from '../types/game'

export const useRoomStore = defineStore('room', () => {
  const rooms = ref<Room[]>([])
  const currentRoom = ref<Room | null>(null)

  async function fetchRooms(zoneId?: string) {
    const url = zoneId ? `/rooms?zone_id=${zoneId}` : '/rooms'
    const { data } = await api.get(url)
    rooms.value = data.rooms ?? []
  }

  async function createRoom(zoneId: string, name: string, mode: GameMode, maxRounds: number = 8) {
    const { data } = await api.post('/rooms', { zone_id: zoneId, name, mode, max_rounds: maxRounds })
    currentRoom.value = data.room
    return data.room_id as string
  }

  async function closeRoom(roomId: string) {
    await api.post(`/rooms/${roomId}/close`)
    currentRoom.value = null
  }

  async function fetchRoom(roomId: string) {
    const { data } = await api.get(`/rooms/${roomId}`)
    currentRoom.value = data
  }

  async function joinRoom(roomId: string) {
    const { data } = await api.post(`/rooms/${roomId}/join`)
    currentRoom.value = data
  }

  async function leaveRoom(roomId: string) {
    await api.post(`/rooms/${roomId}/leave`)
    currentRoom.value = null
  }

  async function addAI(roomId: string): Promise<RoomPlayer> {
    const { data } = await api.post(`/rooms/${roomId}/ai`)
    return data as RoomPlayer
  }

  async function startGame(roomId: string) {
    const { data } = await api.post(`/rooms/${roomId}/start`)
    return data
  }

  async function ready(roomId: string, ready: boolean) {
    await api.post(`/rooms/${roomId}/ready`, { ready })
  }

  function clearCurrentRoom() {
    currentRoom.value = null
  }

  return { rooms, currentRoom, fetchRooms, createRoom, fetchRoom, joinRoom, leaveRoom, addAI, ready, startGame, clearCurrentRoom, closeRoom }
})
