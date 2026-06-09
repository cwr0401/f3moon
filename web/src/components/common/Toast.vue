<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ modelValue?: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const visible = ref(false)
const message = ref('')

watch(() => props.modelValue, (val) => {
  if (val) {
    message.value = val
    visible.value = true
    setTimeout(() => {
      visible.value = false
      emit('update:modelValue', '')
    }, 3000)
  }
})
</script>

<template>
  <Transition name="toast">
    <div
      v-if="visible"
      class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 px-5 py-3 rounded-lg bg-bg-card border border-amber-900/40 shadow-xl text-gray-100"
    >
      {{ message }}
    </div>
  </Transition>
</template>

<style scoped>
.toast-enter-active { transition: all 0.3s ease-out; }
.toast-leave-active { transition: all 0.2s ease-in; }
.toast-enter-from { opacity: 0; transform: translate(-50%, 1rem); }
.toast-leave-to { opacity: 0; transform: translate(-50%, -0.5rem); }
</style>
