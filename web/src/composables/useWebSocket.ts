import { ref, onMounted, onUnmounted, watch, type Ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useGameStore } from '../stores/game'
import type { WSMessage } from '../types/ws'

export function useWebSocket(gameId: Ref<string>) {
  const status = ref<'connecting' | 'open' | 'closed'>('closed')
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0

  function getWsUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = import.meta.env.DEV ? 'localhost:8080' : window.location.host
    return `${protocol}//${host}`
  }

  function connect() {
    if (!gameId.value) return
    const auth = useAuthStore()
    if (!auth.token) return

    disconnect()
    status.value = 'connecting'

    const baseUrl = getWsUrl()
    ws = new WebSocket(`${baseUrl}/ws?token=${auth.token}&game_id=${gameId.value}`)

    ws.onopen = () => {
      status.value = 'open'
      reconnectAttempts = 0
    }

    ws.onmessage = (e: MessageEvent) => {
      try {
        const msg: WSMessage = JSON.parse(e.data as string)
        const gameStore = useGameStore()
        gameStore.applyNotify(msg)
      } catch {
        console.error('Failed to parse WS message')
      }
    }

    ws.onclose = () => {
      status.value = 'closed'
      const delay = Math.min(3000 * Math.pow(2, reconnectAttempts), 30000)
      reconnectAttempts++
      reconnectTimer = setTimeout(connect, delay)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
    status.value = 'closed'
  }

  function send(data: unknown) {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data))
    }
  }

  onMounted(connect)
  onUnmounted(disconnect)
  watch(gameId, () => { connect() })

  return { status, send, disconnect }
}
