import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import api from '../composables/useApi'
import type { GameState, Player } from '../types/game'
import { GamePhase } from '../types/game'
import type { Tile, TileName } from '../types/tile'
import type { WSMessage } from '../types/ws'
import { useAuthStore } from './auth'

export const useGameStore = defineStore('game', () => {
  const gameState = ref<GameState | null>(null)
  const selectedTileId = ref<number | null>(null)
  const cutPlayerIndex = ref<number | null>(null)

  const currentPhase = computed(() => gameState.value?.phase ?? GamePhase.PhaseWaiting)
  const myPlayer = computed<Player | null>(() => {
    const auth = useAuthStore()
    if (!auth.user || !gameState.value) return null
    return gameState.value.players.find(p => p?.id === auth.user!.id) ?? null
  })
  const myHand = computed<Tile[]>(() => myPlayer.value?.hand ?? [])
  const myOpenCombs = computed(() => myPlayer.value?.open_combs ?? [])
  const isMyTurn = computed(() => {
    if (!gameState.value || !myPlayer.value) return false
    return gameState.value.current_turn === gameState.value.players.indexOf(myPlayer.value)
  })
  const isCutPlayer = computed(() => {
    if (!gameState.value || cutPlayerIndex.value === null) return false
    return myPlayer.value?.role === cutPlayerIndex.value
  })

  async function fetchGame(gameId: string) {
    const { data } = await api.get(`/games/${gameId}`)
    gameState.value = data
  }

  async function cut(position: number) {
    await api.post(`/games/${gameState.value!.id}/cut`, { position })
    await fetchGame(gameState.value!.id)
  }

  async function deal() {
    await api.post(`/games/${gameState.value!.id}/deal`)
    await fetchGame(gameState.value!.id)
  }

  async function tong(tileName: TileName, tongSize: number, skip: boolean) {
    await api.post(`/games/${gameState.value!.id}/tong`, { tile_name: tileName, tong_size: tongSize, skip })
    await fetchGame(gameState.value!.id)
  }

  async function draw() {
    await api.post(`/games/${gameState.value!.id}/draw`)
  }

  async function discard(tileId: number) {
    await api.post(`/games/${gameState.value!.id}/discard`, { tile_id: tileId })
  }

  async function pair(tileId: number, pairSize: number) {
    await api.post(`/games/${gameState.value!.id}/pair`, { tile_id: tileId, pair_size: pairSize })
  }

  async function ganta() {
    await api.post(`/games/${gameState.value!.id}/ganta`)
  }

  async function win() {
    await api.post(`/games/${gameState.value!.id}/win`)
  }

  async function pass() {
    await api.post(`/games/${gameState.value!.id}/pass`)
  }

  async function dangJing(jing: TileName) {
    await api.post(`/games/${gameState.value!.id}/dang-jing`, { jing })
  }

  function applyNotify(msg: WSMessage) {
    if (!gameState.value) return
    if (msg.game_id !== gameState.value.id) return

    switch (msg.type) {
      case 'game:phase':
        gameState.value.phase = (msg.data as { phase: GamePhase }).phase
        break
      case 'game:turn':
        gameState.value.current_turn = (msg.data as { current_turn: number }).current_turn
        break
      case 'game:draw':
      case 'game:discard':
      case 'game:pair':
      case 'game:ganta':
      case 'game:win':
      case 'game:huang':
        // Full state refresh for complex events
        fetchGame(gameState.value.id)
        break
      case 'player:hand': {
        const handData = msg.data as { player_id: string; hand: Tile[] }
        const p = gameState.value.players.find(pl => pl?.id === handData.player_id)
        if (p) p.hand = handData.hand
        break
      }
      case 'game:cut-wait': {
        const cutData = msg.data as { cut_player_index: number }
        cutPlayerIndex.value = cutData.cut_player_index
        gameState.value.phase = GamePhase.PhaseCut
        break
      }
      case 'game:dealt':
        fetchGame(gameState.value.id)
        break
      case 'game:tong-ask':
        gameState.value.phase = GamePhase.PhaseTongAsk
        break
      case 'game:tong-turn': {
        const td = msg.data as { tong_current: number; tong_order: number[] }
        gameState.value.tong_current = td.tong_current
        gameState.value.tong_order = td.tong_order
        break
      }
      case 'game:start':
      case 'game:check':
        fetchGame(gameState.value.id)
        break
    }
  }

  function selectTile(tileId: number | null) {
    selectedTileId.value = tileId
  }

  function reset() {
    gameState.value = null
    selectedTileId.value = null
    cutPlayerIndex.value = null
  }

  return {
    gameState, selectedTileId, currentPhase, myPlayer, myHand, myOpenCombs, isMyTurn, isCutPlayer,
    fetchGame, cut, deal, tong, draw, discard, pair, ganta, win, pass, dangJing,
    applyNotify, selectTile, reset,
  }
})
