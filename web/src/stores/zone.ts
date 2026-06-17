import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '../composables/useApi'
import type { GameZone } from '../types/zone'

export const useZoneStore = defineStore('zone', () => {
  const zones = ref<GameZone[]>([])

  async function fetchZones() {
    const { data } = await api.get('/zones')
    zones.value = data.zones ?? []
  }

  return { zones, fetchZones }
})
