<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../../stores/room'
import { GameMode } from '../../types/game'

const props = defineProps<{
  zoneId: string | null
}>()
const emit = defineEmits<{ close: [] }>()
const router = useRouter()
const roomStore = useRoomStore()

const name = ref('')
const mode = ref<GameMode>(GameMode.Mode3Player)
const maxRounds = ref(8)
const loading = ref(false)
const error = ref('')

async function handleCreate() {
  if (!props.zoneId) {
    error.value = '请先选择游戏区'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const roomId = await roomStore.createRoom(props.zoneId, name.value || '花牌局', mode.value, maxRounds.value)
    emit('close')
    router.push(`/room/${roomId}`)
  } catch {
    error.value = '创建房间失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" @click.self="emit('close')">
    <div class="bg-bg-card rounded-xl p-6 w-full max-w-md border border-amber-900/30 shadow-2xl">
      <h2 class="text-xl font-bold text-accent mb-4">创建房间</h2>

      <form @submit.prevent="handleCreate" class="space-y-4">
        <div>
          <label class="block text-sm text-gray-400 mb-1">房间名称</label>
          <input
            v-model="name"
            type="text"
            placeholder="花牌局"
            class="w-full px-4 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/50"
          />
        </div>

        <div>
          <label class="block text-sm text-gray-400 mb-1">游戏模式</label>
          <div class="flex gap-3">
            <button
              type="button"
              @click="mode = GameMode.Mode3Player"
              :class="['flex-1 py-2.5 rounded-lg border transition-colors', mode === GameMode.Mode3Player ? 'bg-accent/20 border-accent text-accent' : 'bg-bg-input border-amber-900/30 text-gray-400']"
            >
              3人定庄
            </button>
            <button
              type="button"
              @click="mode = GameMode.Mode4Player"
              :class="['flex-1 py-2.5 rounded-lg border transition-colors', mode === GameMode.Mode4Player ? 'bg-accent/20 border-accent text-accent' : 'bg-bg-input border-amber-900/30 text-gray-400']"
            >
              4人含歇家
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm text-gray-400 mb-1">最大局数</label>
          <input
            v-model.number="maxRounds"
            type="number"
            min="1"
            max="100"
            class="w-full px-4 py-2.5 rounded-lg bg-bg-input border border-amber-900/30 text-gray-100 focus:outline-none focus:ring-2 focus:ring-accent/50"
          />
        </div>

        <div v-if="error" class="text-red-400 text-sm bg-red-900/20 px-3 py-2 rounded">{{ error }}</div>

        <div class="flex gap-3 pt-2">
          <button type="button" @click="emit('close')" class="flex-1 py-2.5 rounded-lg bg-bg-input text-gray-400 hover:bg-amber-900/20 transition-colors">
            取消
          </button>
          <button type="submit" :disabled="loading" class="flex-1 py-2.5 rounded-lg bg-accent text-bg-dark font-bold hover:bg-accent-dim transition-colors disabled:opacity-50">
            {{ loading ? '创建中...' : '创建' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
